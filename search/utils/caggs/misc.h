/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

#ifndef __CAGGS_MISC_H__
#define __CAGGS_MISC_H__

#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>


/**
 * @brief Parse length in bytes, KBytes or Megabytes.
 * @param[in] str String to parse.
 * @param[out] len Length in bytes.
 * @return Zero on success.
 */
int parse_len(const char *str, int64_t *len);



/**
 * @brief Get current time.
 * @return Current timestamp, microseconds
 *         or `0` in case of any errror.
 */
int64_t get_time();


/**
 * @brief Get current time in seconds.
 */
static inline double get_time_sec()
{
    return get_time() * 1e-6;
}


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
