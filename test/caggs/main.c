/** @file
 * @brief Main application.
 */
#include "conf.h"
#include "misc.h"

#include <signal.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>
#include <math.h>

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
 * Tries to parse the INDEX line in the following format:
 * `filename,offset,length,fuzziness`.
 * The `filename`, `offset` and `fuzziness` are ignored.
 * So just `length` is parsed to the `data_len`.
 *
 * @param[in] idx_beg Begin of INDEX.
 * @param[in] idx_end End of INDEX.
 * @param[out] data_len Length of DATA in bytes.
 * @return Zero on success.
 */
static int parse_index(const uint8_t *idx_beg, const uint8_t *idx_end, uint64_t *data_len)
{
    const int COMMA = ',';

    // find ",fuzziness"
    const uint8_t *c4 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                idx_end - idx_beg);
    if (!c4)
        return -1; // no ",fuzziness" found

    // find ",length"
    const uint8_t *c3 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                c4 - idx_beg);
    if (!c3)
        return -2; // no ",length" found

    uint8_t *end = 0;
    *data_len = strtoull((const char*)c3+1, // +1 to skip comma
                         (char**)&end, 10);
    if (end != c4)
        return -3; // failed to parse length

    return 0; // OK
}


// TODO: abstract classes, dedicated files, etc
struct Stat {
    uint64_t count;
    double sum, sum2;
    double min, max;
};

// calculate statistics
void stat_add(struct Stat *s, double x)
{
    if (!s->count || x < s->min)
        s->min = x;
    if (!s->count || x > s->max)
        s->max = x;
    s->sum += x;
    s->sum2 += x*x;
    s->count += 1;
}

// print statistics to STDOUT
void stat_print(const struct Stat *s)
{
    if (s->count)
    {
        const double avg = s->sum/s->count;
//        const double var = s->sum2/s->count - avg*avg;
//        const double stdev = sqrt(var);
//        const double sigma = 2.0;

        vlog("{\"avg\":%f, \"sum\":%f, \"min\":%f, \"max\":%f, \"count\":%llu}\n",
             avg, s->sum, s->min, s->max, s->count);
    }
    else
    {
        vlog("{\"avg\":null, \"sum\":0, \"min\":null, \"max\":null, \"count\":0}\n");
    }
}

struct Stat g_stat;


/**
 * @brief Process DATA record.
 * @param[in] cfg Application configuration.
 * @param[in] dat Begin of DATA.
 * @param[in] len Length of DATA in bytes.
 * @return Zero on success.
 */
static int process_record(const struct Conf *cfg, const uint8_t *beg, const uint8_t *end)
{
    (void)cfg; // not used yet

//    printf("  RECORD[%llu]:", len);
//    for (; len > 0; --len)
//        printf("%c", *dat++);
//    printf("\n");

    const char *field = "foo\"";

    while ((end - beg) > 0)
    {
        const uint8_t *f = (const uint8_t*)memchr(beg, field[0], end - beg);
        if (!f)
            return -1; // field not found

        if (0 != memcmp(f, field, 4))
        {
            beg = f+1;
            continue; // try again
        }

        beg = f+4;
        while ((end - beg) > 0)
        {
            if (*beg++ == ':')
                break;
        }

        if (!(end - beg))
            return -2; // no data found

        double x = strtod((const char*)beg, NULL);
        // vlog(" %g ", x);
        stat_add(&g_stat, x);
        break; // done
    }

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
                   const uint8_t *idx_beg, const uint8_t *idx_end,
                   const uint8_t *dat_beg, const uint8_t *dat_end)
{
    // remove DATA header
    if ((ptrdiff_t)cfg->header_len <= (dat_end - dat_beg))
        dat_beg += cfg->header_len;
    else
    {
        verr("ERROR: no DATA available (%d) to skip header (%d)\n",
             (dat_end - dat_beg), cfg->header_len);
        return -1; // failed
    }

    // remove DATA footer
    if ((off_t)cfg->footer_len <= (dat_end - dat_beg))
        dat_end -= cfg->footer_len;
    else
    {
        verr("ERROR: no DATA available (%d) to skip footer (%d)\n",
             dat_end - dat_beg, cfg->footer_len);
        return -1; // failed
    }

    // read INDEX line by line
    uint64_t count = 0;
    while ((idx_end - idx_beg) > 0)
    {
        // try to find the NEWLINE '\n' character
        const uint8_t *eol = (const uint8_t*)memchr(idx_beg, '\n',
                                                    idx_end - idx_beg);
        const uint8_t *next = eol ? (eol + 1) : idx_end;

        uint64_t d_len = 0;
        int res = parse_index(idx_beg, next, &d_len);
        if (res != 0)
        {
            verr("ERROR: failed to parse INDEX: %d\n", res); // TODO: add "at" information from idx_p
            return -2; // failed
        }

        // TODO: concurrency!!!
        if ((ptrdiff_t)d_len <= (dat_end - dat_beg))
        {
            int res = process_record(cfg, dat_beg,
                                     dat_beg + d_len);
            if (res != 0)
            {
                verr("ERROR: failed to process RECORD: %s\n", res);// TODO: add "at" information from dat_p
                return -3; // failed
            }

            // go to next record...
            dat_beg += d_len + cfg->delim_len;
        }
        else
        {
            verr("ERROR: no DATA found\n"); // TODO: add "at" information from dat_p
            return -4; // failed
        }

        // go to next line
        idx_beg = next;
        count += 1;

        if (g_stopped != 0)
            break;
    }

    vlog2("total records processed: %llu\n", count);
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
    do_work(&cfg, (const uint8_t*)idx_p, (const uint8_t*)idx_p + idx_stat.st_size,
            (const uint8_t*)dat_p, (const uint8_t*)dat_p + dat_stat.st_size);

    // print global statistics
    stat_print(&g_stat);

    // release resources
    munmap(idx_p, idx_stat.st_size);
    munmap(dat_p, dat_stat.st_size);
    close(dat_fd);
    close(idx_fd);
    conf_free(&cfg);

    return 0;
}
