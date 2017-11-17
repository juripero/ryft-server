/** @file
 * @brief Main application.
 */
#include "conf.h"
#include "misc.h"

#include <signal.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <errno.h>

#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <unistd.h>
#include <fcntl.h>


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


/**
 * @brief Parse the INDEX information.
 *
 * Tries to parse the line in the following format:
 * `filename,offset,length,fuzziness`.
 * The `filename`, `offset` and `fuzziness` are ignored.
 *
 * @param[in] idx The begin of INDEX line.
 * @param[in] idx_len The length of INDEX line in bytes.
 * @param[out] data_len The DATA record length in bytes.
 * @return Zero on success.
 */
static int parse_index(const uint8_t *idx, uint64_t idx_len, uint64_t *data_len)
{
    const uint8_t *c4 = (const uint8_t*)memrchr(idx, ',', idx_len);
    if (!c4)
        return -1; // no ",fuzziness" found

    const uint8_t *c3 = (const uint8_t*)memrchr(idx, ',', c4-idx);
    if (!c3)
        return -2; // no ",length" found

    char *end = 0;
    *data_len = strtoull((const char*)c3+1, &end, 10);
    if ((const uint8_t*)end != c4)
        return -3; // failed to parse length

    return 0; // OK
}


/**
 * @brief Do the work.
 * @param[in] cfg Application configuration.
 * @param[in] idx_p The begin of INDEX file.
 * @param[in] idx_len The length of INDEX file in bytes.
 * @param[in] dat_p The begin of DATA file.
 * @param[in] dat_len The length of DATA file in bytes.
 * @return Zero on success.
 */
static int do_work(const struct Conf *cfg,
                   const uint8_t *idx_p, uint64_t idx_len,
                   const uint8_t *dat_p, uint64_t dat_len)
{
    // remove DATA header
    if (cfg->header_len <= dat_len)
    {
        dat_p += cfg->header_len;
        dat_len -= cfg->header_len;
    }
    else
    {
        verr("ERROR: no DATA avaialbe (%d) to skip header (%d)\n",
             dat_len, cfg->header_len);
        return -1; // failed
    }

    // remove DATA footer
    if (cfg->footer_len <= dat_len)
    {
        dat_len -= cfg->footer_len;
    }
    else
    {
        verr("ERROR: no DATA avaialbe (%d) to skip footer (%d)\n",
             dat_len, cfg->footer_len);
        return -1; // failed
    }

    // read INDEX line by line
    uint64_t idx_count = 0;
    while (idx_len > 0)
    {
        // try to find the NEWLINE '\n' character
        const uint8_t *eol = (const uint8_t*)memchr(idx_p, '\n', idx_len);
        const uint64_t len = eol ? (uint64_t)(eol - idx_p + 1) : idx_len;

        uint64_t data_len = 0;
        int res = parse_index(idx_p, len, &data_len);
        if (res != 0)
        {
            verr("ERROR: failed to parse INDEX: %d\n", res); // TODO: add "at" information from idx_p
            return -2; // failed
        }

        vlog("  new INDEX#%llu (%llu bytes) with %llu bytes of DATA\n",
             idx_count, len, data_len);

        // go to next line
        idx_p += len;
        idx_len -= len;
        idx_count += 1;

        if (g_stopped != 0)
            break;
    }

    vlog2("total INDEXes: %d\n", idx_count);

    // TODO: read DATA record-by-record
    // TODO: process records

    return 0; // OK
}


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

    struct Conf cfg;
    memset(&cfg, 0, sizeof(cfg));
    if (!!conf_parse(&cfg, argc, argv))
        return -1;

    // print current configuration
    if (verbose >= 3)
        conf_print(&cfg);

    // try to open INDEX file
    vlog2("opening INDEX file: %s\n", cfg.idx_path);
    int idx_fd = open(cfg.idx_path, O_RDONLY/*|O_LARGEFILE*/);
    if (idx_fd < 0)
    {
        verr("ERROR: failed to open INDEX file: %s\n",
             strerror(errno));
        return -1;
    }
    // and get INDEX file size
    struct stat idx_stat;
    memset(&idx_stat, 0, sizeof(idx_stat));
    if (!!fstat(idx_fd, &idx_stat))
    {
        verr("ERROR: failed to stat INDEX file: %s\n",
             strerror(errno));
        return -1;
    }
    vlog2("        INDEX file: #%d (%d bytes)\n",
          idx_fd, idx_stat.st_size);

    // try to open DATA file
    vlog2("opening  DATA file: %s\n", cfg.dat_path);
    int dat_fd = open(cfg.dat_path, O_RDONLY/*|O_LARGEFILE*/);
    if (dat_fd < 0)
    {
        verr("ERROR: failed to open DATA file: %s\n",
             strerror(errno));
        return -1;
    }
    // and get DATA file size
    struct stat dat_stat;
    memset(&dat_stat, 0, sizeof(dat_stat));
    if (!!fstat(dat_fd, &dat_stat))
    {
        verr("ERROR: failed to stat DATA file: %s\n",
             strerror(errno));
        return -1;
    }
    vlog2("         DATA file: #%d (%d bytes)\n",
          dat_fd, dat_stat.st_size);

    // TODO: do memory mapping part-by-part
    // let say 64MB per each part.

    // do memory mapping
    void *idx_p = mmap(0, idx_stat.st_size, PROT_READ, MAP_SHARED, idx_fd, 0);
    if (MAP_FAILED == idx_p)
    {
        verr("ERROR: failed to map INDEX file: %s\n",
             strerror(errno));
        return -1;
    }
    void *dat_p = mmap(0, dat_stat.st_size, PROT_READ, MAP_SHARED, dat_fd, 0);
    if (MAP_FAILED == dat_p)
    {
        verr("ERROR: failed to map DATA file: %s\n",
             strerror(errno));
        return -1;
    }

    // do actual processing
    do_work(&cfg, (const uint8_t*)idx_p, idx_stat.st_size,
            (const uint8_t*)dat_p, dat_stat.st_size);

    // release resources
    munmap(idx_p, idx_stat.st_size);
    munmap(dat_p, dat_stat.st_size);
    close(dat_fd);
    close(idx_fd);
    conf_free(&cfg);

    return 0;
}
