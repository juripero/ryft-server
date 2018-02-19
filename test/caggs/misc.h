#ifndef __CAGGS_MISC_H__
#define __CAGGS_MISC_H__

#include <stdarg.h>
#include <stdio.h>


/**
 * @brief Parse length in bytes, KBytes or Megabytes.
 * @param[in] str String to parse.
 * @param[out] len Length in bytes.
 * @return Zero on success.
 */
int parse_len(const char *str, int *len);


/**
 * @brief Verbose level.
 *
 * - `0` - be quiet
 * - `1` - show important messages
 * - `2` - show detailed messages
 * - `3` - show all (trace) messages
 */
extern int verbose;


/**
 * @brief Print message to the STDOUT and flush the stream.
 */
static inline void vlog(const char *msg, ...)
{
    va_list args;
    va_start(args, msg);
    vfprintf(stdout, msg, args);
    va_end(args);

    fflush(stdout);
}


/**
 * @brief Print message to the STDERR and flush the stream.
 */
static inline void verr(const char *msg, ...)
{
    va_list args;
    va_start(args, msg);
    vfprintf(stderr, msg, args);
    va_end(args);

    fflush(stderr);
}


// print INFO messages
#define vlog1(...) if (verbose < 1) {} else vlog(__VA_ARGS__)

// print DETAIL messages
#define vlog2(...) if (verbose < 2) {} else vlog(__VA_ARGS__)

// print TRACE messages
#define vlog3(...) if (verbose < 3) {} else vlog(__VA_ARGS__)


// print INFO error messages
#define verr1(...) if (verbose < 1) {} else verr(__VA_ARGS__)

// print DETAIL error messages
#define verr2(...) if (verbose < 2) {} else verr(__VA_ARGS__)

// print TRACE error messages
#define verr3(...) if (verbose < 3) {} else verr(__VA_ARGS__)


#endif // __CAGGS_MISC_H__
