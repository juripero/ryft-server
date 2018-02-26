/** @file
 * @brief Main application.
 */
#include "conf.h"
#include "misc.h"
#include "json.h"
#include "stat.h"
#include "proc.h"

#include <signal.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>
#include <math.h>

#include <inttypes.h>
#include <stdatomic.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <unistd.h>
#include <fcntl.h>

#include <pthread.h>


// some global variables
static int volatile g_stopped = 0;


/**
 * @brief Handle system signals.
 * @param[in] signo The signal number.
 */
static void signal_handler(int signo)
{
    switch (signo)
    {
    case SIGINT:
        vlog1("SIGINT received, stopping...\n");
        g_stopped = 1; // stop main loop
        break;

    case SIGTERM:
        vlog1("SIGTERM received, stopping...\n");
        g_stopped = 1; // stop main loop
        break;

    default:
        vlog1("%d signal received, do nothing\n", signo);
        break;
    }
}

#ifdef NO_TESTS

/**
 * @brief Application entry point.
 * @param[in] argc Number of command line arguments.
 * @param[in] argv List of command line arguments.
 * @return Zero on success.
 */
int main(int argc, const char *argv[])
{
    // setup signal handlers
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);

    // parse configuration
    struct Conf cfg;
    memset(&cfg, 0, sizeof(cfg));
    if (!!conf_parse(&cfg, argc, argv))
        return -1;

    // print current configuration
    if (verbose >= 3)
        conf_print(&cfg);

    // create processing work
    struct Work *work = work_make(&cfg);
    if (!work)
        return -1;

    // get PAGE_SIZE from system
    const long page_size = sysconf(_SC_PAGE_SIZE);
    if (page_size <= 0)
    {
        verr("ERROR: failed to get page size: %s\n",
             strerror(errno));
        return -1;
    }
    if (page_size & (page_size-1))
    {
        verr("ERROR: page size is not power of two: %ld\n",
             page_size);
        return -1;
    }
    vlog3("page size: %ld\n", page_size);


    // INDEX/DATA file parameters
    struct FileP
    {
        int      fd;    // file descriptor
        uint64_t pos;   // read position
        uint64_t len;   // file length, bytes
    } i_file, d_file;

    if (1) // try to open INDEX file
    {
        vlog2("opening INDEX file: %s\n", cfg.idx_path);
        i_file.fd = open(cfg.idx_path, O_RDONLY/*|O_LARGEFILE*/);
        if (i_file.fd < 0)
        {
            verr("ERROR: failed to open INDEX file: %s\n",
                 strerror(errno));
            return -1;
        }

        // and get INDEX file size
        struct stat s;
        memset(&s, 0, sizeof(s));
        if (!!fstat(i_file.fd, &s))
        {
            verr("ERROR: failed to stat INDEX file: %s\n",
                 strerror(errno));
            return -1;
        }
        i_file.len = s.st_size;
        i_file.pos = 0;

        vlog2("        INDEX file: #%d (%"PRIu64" bytes)\n",
              i_file.fd, i_file.len);
    }

    if (1) // try to open DATA file
    {
        vlog2("opening  DATA file: %s\n", cfg.dat_path);
        d_file.fd = open(cfg.dat_path, O_RDONLY/*|O_LARGEFILE*/);
        if (d_file.fd < 0)
        {
            verr("ERROR: failed to open DATA file: %s\n",
                 strerror(errno));
            return -1;
        }

        // and get DATA file size
        struct stat s;
        memset(&s, 0, sizeof(s));
        if (!!fstat(d_file.fd, &s))
        {
            verr("ERROR: failed to stat DATA file: %s\n",
                 strerror(errno));
            return -1;
        }
        d_file.len = s.st_size;
        d_file.pos = 0; // read position

        vlog2("         DATA file: #%d (%"PRIu64" bytes)\n",
              d_file.fd, d_file.len);
    }


    // memory chunk parameters
    struct ChunkP
    {
        int id;         // identifier (for log/debug purposes)
        uint8_t *base;  // base address (mapped memory)
        uint64_t pos;   // read position
        uint64_t len;   // chunk length, bytes
    } i_buf;

    memset(&i_buf, 0, sizeof(i_buf));
    int d_buf_id = 0; // DATA chunk identifier

    struct RecordRef *records = (struct RecordRef*)malloc(cfg.rec_per_chunk * sizeof(*records));

    d_file.pos += cfg.header_len; // skip DATA header
    while (!g_stopped && d_file.pos < (d_file.len - cfg.footer_len)) // keep in mind DATA footer!
    {
        const uint64_t d_align = d_file.pos & (page_size-1); // DATA aligment
        uint64_t num_of_records = 0; // actual number of record references parsed
        uint64_t data_len = d_align; // corresponding DATA chunk size

        // prepare DATA chunk: parse one or more INDEX chunks until
        // requered number of record references will be parsed
        while (!g_stopped && i_file.pos < i_file.len
            && num_of_records < cfg.rec_per_chunk
            && data_len < cfg.dat_chunk_size)
        {
            if (!i_buf.base) // no valid INDEX chunk, prepare next one
            {
                const uint64_t i_align = i_file.pos & (page_size-1); // INDEX aligment
                const uint64_t rem = i_file.len - (i_file.pos - i_align); // remain INDEX bytes
                const uint64_t len = rem < cfg.idx_chunk_size
                                   ? rem : cfg.idx_chunk_size;

                // do the INDEX mapping
                void *base = mmap(0, len, PROT_READ, MAP_SHARED,
                                  i_file.fd, i_file.pos - i_align);
                if (MAP_FAILED == base)
                {
                    verr("ERROR: failed to map INDEX file: %s\n",
                         strerror(errno));
                    return -1;
                }

                vlog2("new IndexChunk%d of %"PRIu64" bytes (at %"PRIu64"-%"PRIu64"=%"PRIu64") prepared\n",
                      i_buf.id, len, i_file.pos, i_align, i_file.pos - i_align);

                i_buf.base = (uint8_t*)base;
                i_buf.len = len;
                i_buf.pos = i_align;
            }

            // parse indices to record referenecs
            const uint8_t *buf = i_buf.base + i_buf.pos;
            uint64_t i_len = i_buf.len - i_buf.pos;         // remain of INDEX bytes => actual number of INDEX bytes processed
            uint64_t d_len = cfg.dat_chunk_size - data_len; // remain of DATA bytes => actual number of DATA bytes needed
            uint64_t n_rec = cfg.rec_per_chunk - num_of_records;    // remain space => actual number of record references parsed
            const int is_last = (i_file.len - i_file.pos) <= i_len; // last INDEX chunk flag
            const int res = parse_index_chunk(is_last, buf, &i_len,
                                              cfg.delim_len, data_len, &d_len,
                                              records + num_of_records, &n_rec);
            if (res < 0)
            {
                verr("ERROR: failed to parse INDEX file\n");
                return -1;
            }

            vlog2("IndexChunk%d: %"PRIu64" indices, %"PRIu64" DATA bytes, "
                  "%"PRIu64" INDEX bytes, INDEX:[%"PRIu64"..%"PRIu64")\n",
                  i_buf.id, n_rec, d_len, i_len,
                  i_file.pos, i_file.pos + i_len);

            // update current chunk
            i_file.pos += i_len;
            i_buf.pos += i_len;
            data_len += d_len;
            num_of_records += n_rec;
            if (res > 0 || i_buf.pos >= i_buf.len)
            {
                // (res > 0) means that INDEX chunk is fully processed

                // release INDEX chunk
                if (!!munmap((void*)i_buf.base, i_buf.len))
                {
                    verr("ERROR: failed to unmap INDEX file: %s\n",
                         strerror(errno));
                    return -1;
                }

                vlog2("IndexChunk%d of %"PRIu64" bytes released\n",
                      i_buf.id, i_buf.len);

                i_buf.base = 0;
                i_buf.len = 0;
                i_buf.id += 1;
            }
            else
            {
                // current INDEX chunk is not fully processed
                // but DATA chunk is already full
                break; // so we are done for now
            }
        }

        vlog2("DataChunk%d: %"PRIu64" records, %"PRIu64" DATA bytes, DATA:[%"PRIu64"..%"PRIu64")\n",
              d_buf_id, num_of_records, data_len,
              d_file.pos, d_file.pos + data_len);
        if (!data_len|| !num_of_records)
            break;

        // prepare DATA chunk
        void *base = mmap(0, data_len, PROT_READ, MAP_SHARED,
                          d_file.fd, d_file.pos - d_align);
        if (MAP_FAILED == base)
        {
            verr("ERROR: failed to map DATA file: %s\n",
                 strerror(errno));
            return -1;
        }

        // do actual processing
        if (!!work_do(work, (uint8_t*)base,
                      records, num_of_records))
        {
            verr("ERROR: failed to do processing\n");
            return -1;
        }

        // release DATA chunk
        if (!!munmap(base, data_len))
        {
            verr("ERROR: failed to unmap DATA file: %s\n",
                 strerror(errno));
            return -1;
        }

        d_file.pos += data_len - d_align;
        d_buf_id += 1;
    }

    // print results
    work_print(work, stdout);

    close(d_file.fd);
    close(i_file.fd);
    work_free(work);
    conf_free(&cfg);

    return 0;
}

#endif // NO_TESTS
