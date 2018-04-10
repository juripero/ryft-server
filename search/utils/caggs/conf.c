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
    vlog("-X<N>, --concurrency=<N> do processing in N threads (8 by default)\n");
    vlog("--native          use \"native\" output format\n");
    vlog("\n");

    vlog("-i<path>, --index=<path> path to INDEX file\n");
    vlog("-d<path>, --data=<path>  path to DATA file\n");
    vlog("-f<path>, --field=<path> field to access JSON data\n");
    vlog("\n");

    vlog("-H<N>, --header=<N> size of DATA header in bytes\n");
    vlog("-D<N>, --delim=<N>  size of DATA delimiter in bytes\n");
    vlog("-F<N>, --footer=<N> size of DATA footer in bytes\n");
    vlog("\n");

    vlog("-b<N>, --index-chunk=<N> size of INDEX chunk in bytes (512MB by default)\n");
    vlog("-B<N>, --data-chunk=<N>  size of DATA chunk in bytes (512MB by default)\n");
    vlog("-R<N>, --max-records=<N> maximum number of records per DATA chunk (16M by default)\n");
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
    cfg->fields = 0;
    cfg->n_fields = 0;
    cfg->header_len = 0;
    cfg->delim_len = 0;
    cfg->footer_len = 0;
    cfg->idx_chunk_size = 64*(1024*1024);
    cfg->dat_chunk_size = 64*(1024*1024);
    cfg->rec_per_chunk = 16*(1024*1024);
    cfg->concurrency = 8;

    // long options
    struct option opts[] =
    {
        {"index", required_argument, 0, 'i' },
        {"data", required_argument, 0, 'd' },
        {"field", required_argument, 0, 'f' },
        {"header", required_argument, 0, 'H' },
        {"delimiter", required_argument, 0, 'D' },
        {"delim", required_argument, 0, 'D' },
        {"footer", required_argument, 0, 'F' },
        {"index-chunk", required_argument, 0, 'b' },
        {"data-chunk", required_argument, 0, 'B' },
        {"max-records", required_argument, 0, 'R' },
        {"concurrency", required_argument, 0, 'X' },

        { "help", no_argument, 0, 'h' },
        { "version", no_argument, 0, 'V' },
        { "quiet", no_argument, 0, 'q' },
        { "verbose", no_argument, 0, 'v' },
        { "native", no_argument, 0, 'N' },

        { 0, 0, 0, 0 } // EOF
    };

    while (1)
    {
        // parse options
        int res = getopt_long(argc, (char* const*)argv,
                              "i:d:f:H:D:F:b:B:R:X:hVqv", opts, 0);
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

        case 'N': // native format
            // TODO: set corresponding flag
            break;

        case 'i': // INDEX file path
            cfg->idx_path = optarg; // TODO: do we need to copy?
            break;

        case 'd': // DATA file path
            cfg->dat_path = optarg; // TODO: do we need to copy?
            break;

        case 'f': // field
        {
            const char **fields = (const char**)realloc(cfg->fields,
                                                        (cfg->n_fields+1)*sizeof(cfg->fields[0]));
            if (!fields)
            {
                verr("failed to allocate memory for fields\n");
                return -1;
            }
            cfg->fields = fields;
            cfg->fields[cfg->n_fields++] = optarg; // TODO: do we need to copy?
        } break;

        case 'H': // header length
        {
            int64_t n = 0;
            if (!!parse_len(optarg, &n))
            {
                verr("failed to parse header length: %s\n", optarg);
                return -1; // failed

            }
            if (n < 0)
            {
                verr("invalid header length: cannot be negative\n");
                return -1; // failed
            }
            cfg->header_len = n;
        } break;

        case 'D': // delimiter length
        {
            int64_t n = 0;
            if (!!parse_len(optarg, &n))
            {
                verr("failed to parse delimiter length: %s\n", optarg);
                return -1; // failed

            }
            if (n < 0)
            {
                verr("invalid delimiter length: cannot be negative\n");
                return -1; // failed
            }
            cfg->delim_len = n;
        } break;

        case 'F': // footer length
        {
            int64_t n = 0;
            if (!!parse_len(optarg, &n))
            {
                verr("failed to parse footer length: %s\n", optarg);
                return -1; // failed

            }
            if (n < 0)
            {
                verr("invalid footer length: cannot be negative\n");
                return -1; // failed
            }

            cfg->footer_len = n;
        } break;

        case 'b': // INDEX chunk size
        {
            int64_t n = 0;
            if (!!parse_len(optarg, &n))
            {
                verr("failed to parse INDEX chunk size: %s\n", optarg);
                return -1; // failed

            }
            if (n < 1*1024*1024)
            {
                verr("invalid INDEX chunk size: cannot be less than 1MB\n");
                return -1; // failed
            }
            cfg->idx_chunk_size = n;
        } break;

        case 'B': // DATA chunk size
        {
            int64_t n = 0;
            if (!!parse_len(optarg, &n))
            {
                verr("failed to parse DATA chunk size: %s\n", optarg);
                return -1; // failed

            }
            if (n < 1*1024*1024)
            {
                verr("invalid DATA chunk size: cannot be less than 1MB\n");
                return -1; // failed
            }
            cfg->dat_chunk_size = n;
        } break;

        case 'R': // maximum records per chunk
        {
            int64_t n = 0;
            if (!!parse_len(optarg, &n)) // use G, M, K suffixes as for bytes!
            {
                verr("failed to parse records per chunk: %s\n", optarg);
                return -1; // failed

            }
            if (n < 1000)
            {
                verr("invalid records per chunk: cannot be less than 1000\n");
                return -1; // failed
            }
            cfg->rec_per_chunk = n;
        } break;

        case 'X': // concurrency
            cfg->concurrency = strtol(optarg, NULL, 0);
            if (cfg->concurrency < 0 || 64 < cfg->concurrency)
            {
                verr("concurrency \"%d\" is out of range [0..64]\n", cfg->concurrency);
                return -1; // failed
            }
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

    if (!cfg->n_fields)
    {
        verr("ERROR: at least one FIELD should be provided\n");
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
    if (cfg->fields)
        free(cfg->fields);
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
         (int)cfg->header_len,
         (int)cfg->delim_len,
         (int)cfg->footer_len);
    vlog("fields: [");
    for (int i = 0; i < cfg->n_fields; ++i)
        vlog("%s%s", i?", ":"", cfg->fields[i]);
    vlog("]\n");
    vlog("INDEX chunk: %.3gMB\n",
         cfg->idx_chunk_size/(1024*1024.0));
    vlog(" DATA chunk: %.3gMB (%.3gM records maximum)\n",
         cfg->dat_chunk_size/(1024*1024.0),
         cfg->rec_per_chunk/(1024*1024.0));
    vlog("concurrency: x%d\n", cfg->concurrency);
    vlog("  verbosity: %d\n", verbose);
}
