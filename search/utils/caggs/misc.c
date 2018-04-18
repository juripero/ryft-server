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

#include "misc.h"

#include <sys/time.h>
#include <stdlib.h>
#include <string.h>

/*
 * be quiet by default.
 */
int verbose = 0;


/*
 * parse_len() implementation.
 */
int parse_len(const char *str, int64_t *len)
{
    if (!str)
        return -1; // no string provided

    // parse bytes
    char *end = 0;
    double b = strtod(str, &end);

    double x; // scale
    if (0 == strcasecmp(end, "G") || 0 == strcasecmp(end, "GB"))
        x = 1024*1024*1024;
    else if (0 == strcasecmp(end, "M") || 0 == strcasecmp(end, "MB"))
        x = 1024*1024;
    else if (0 == strcasecmp(end, "K") || 0 == strcasecmp(end, "KB"))
        x = 1024;
    else if (0 == strcasecmp(end, "B") || 0 == strcmp(end, ""))
        x = 1;
    else
        return -2; // suffix is unknown!

    // save length (bytes)
    if (len) *len = (b*x + 0.5); // round

    return 0; // OK
}


/*
 * get_time() implementation.
 */
int64_t get_time()
{
    struct timeval tv;
    if (!!gettimeofday(&tv, 0))
    {
        // TODO: report error?
        return 0;
    }

    // convert time value to microseconds
    return tv.tv_sec*(int64_t)1000000 + tv.tv_usec;
}
