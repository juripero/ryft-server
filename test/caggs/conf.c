/** @file
 * @brief Application configuration.
 */
#include "conf.h"
#include "misc.h"

#include <string.h>
#include <unistd.h>
#include <getopt.h>
#include <stdlib.h>

/**
 * @brief Print usage information to STDOUT.
 */
static void usage(void)
{
    vlog("Calculate aggregations: caggs [options]\n");
#if defined(CAGGS_VERSION)
    vlog("(version: %s)\n", CAGGS_VERSION);
#endif // CAGGS_VERSION
    vlog("\n");

    vlog("-h, --help        print this short help\n");
    vlog("-V, --version     print version if available\n");
    vlog("-q, --quiet       be quiet, disable verbose mode\n");
    vlog("-v, --verbose     enable verbose mode (also -vv and -vvv)\n");
    vlog("-P<N>, --concurrency=<N> do processing in N threads\n");
    vlog("\n");

    vlog("-i<path>, --index=<path> path to INDEX file\n");
    vlog("-d<path>, --data=<path>  path to DATA file\n");
    vlog("\n");

    vlog("-H<N>, --header=<N> size of DATA header in bytes\n");
    vlog("-D<N>, --delim=<N>  size of DATA delimiter in bytes\n");
    vlog("-F<N>, --footer=<N> size of DATA footer in bytes\n");
    vlog("\n");

    vlog("The following signals are handled:\n");
    vlog("  SIGINT, SIGTERM - stop the tool\n");
}


/*
 * conf_parse() implementation.
 */
int conf_parse(struct Conf *cfg, int argc, const char *argv[])
{
    // default options
    cfg->header_len = 0;
    cfg->delim_len = 0;
    cfg->footer_len = 0;
    cfg->concurrency = 8;

    // long options
    struct option opts[] =
    {
        {"index", required_argument, 0, 'i' },
        {"data", required_argument, 0, 'd' },
        {"header", no_argument, 0, 'H' },
        {"delimiter", no_argument, 0, 'D' },
        {"delim", no_argument, 0, 'D' },
        {"footer", no_argument, 0, 'F' },
        {"concurrency", no_argument, 0, 'P' },

        { "help", no_argument, 0, 'h' },
        { "version", no_argument, 0, 'V' },
        { "quiet", no_argument, 0, 'q' },
        { "verbose", no_argument, 0, 'v' },

        { 0, 0, 0, 0 } // EOF
    };

    while (1)
    {
        // parse options
        int res = getopt_long(argc, (char* const*)argv,
                              "i:d:H:D:F:P:hVqv", opts, 0);
        if (res < 0)
            break; // done

        switch (res)
        {
        case '?':
            usage();
            return -1; // failed

        case 'h': // show usage
            usage();
            return 1; // stop

        case 'V': // show version
#if defined(CAGGS_VERSION)
            vlog("version: %s\n", CAGGS_VERSION);
#endif // CAGGS_VERSION
#if defined(CAGGS_GITHASH)
            vlog("githash: %s\n", CAGGS_GITHASH);
#endif // CAGGS_GITHASH
#if defined(CAGGS_BUILDTIME)
            vlog("build: %s\n", CAGGS_BUILDTIME);
#endif // CAGGS_BUILDTIME
            return 1; // stop

        case 'q': // be quiet
            verbose = 0;
            break;

        case 'v': // tell more and more...
            verbose += 1;
            break;

        case 'i': // INDEX file path
            cfg->idx_path = optarg; // TODO: do we need to copy?
            break;

        case 'd': // DATA file path
            cfg->dat_path = optarg; // TODO: do we need to copy?
            break;

        case 'H': // header length
            cfg->header_len = strtoul(optarg, NULL, 0); // TODO: check errors
            break;

        case 'D': // delimiter length
            cfg->delim_len = strtoul(optarg, NULL, 0); // TODO: check errors
            break;

        case 'F': // footer length
            cfg->footer_len = strtoul(optarg, NULL, 0); // TODO: check errors
            break;

        case 'P': // concurrency
            cfg->concurrency = strtoul(optarg, NULL, 0); // TODO: check errors
            break;

        default:
            return -1; // failed
        }
     }

    if (!cfg->idx_path)
    {
        verr("ERROR: no INDEX path provided\n");
        // usage();
        return -1; // failed
    }

    if (!cfg->dat_path)
    {
        verr("ERROR: no DATA path provided\n");
        // usage();
        return -1; // failed
    }

    return 0; // OK
}


/*
 * conf_free() implementation.
 */
int conf_free(struct Conf *cfg)
{
    (void)cfg;
    return 0; // OK
}


/*
 * conf_print() implementation.
 */
void conf_print(const struct Conf *cfg)
{
#if defined(CAGGS_VERSION)
    vlog("tool version: %s\n", CAGGS_VERSION);
#endif // CAGGS_VERSION
    vlog("INDEX: %s\n DATA: %s (%d/%d/%d)\n",
         cfg->idx_path,
         cfg->dat_path,
         cfg->header_len,
         cfg->delim_len,
         cfg->footer_len);
    vlog("concurrency: x%d\n", cfg->concurrency);
    vlog("  verbosity: %d\n", verbose);
}
