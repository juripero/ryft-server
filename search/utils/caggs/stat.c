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

#include "stat.h"

#include <stdlib.h>
#include <string.h>


/*
 * stat_init() implementation.
 */
void stat_init(struct Stat *s)
{
    memset(s, 0, sizeof(*s));
}


/*
 * stat_make() implementation.
 */
struct Stat* stat_make(void)
{
    struct Stat *s = (struct Stat*)malloc(sizeof(*s));
    if (!s)
        return 0; // failed

    stat_init(s);
    return s;
}


/*
 * stat_clone() implementation.
 */
struct Stat* stat_clone(const struct Stat *base)
{
    struct Stat *s = (struct Stat*)malloc(sizeof(*s));
    if (!s)
        return 0; // failed

    memcpy(s, base, sizeof(*s));
    return s;
}


/*
 * stat_free() implementation.
 */
void stat_free(struct Stat *s)
{
    free(s);
}


/*
 * stat_add() implementation.
 */
void stat_add(struct Stat *s, double x)
{
    // minimum
    if (!s->count || x < s->min)
        s->min = x;

    // maximum
    if (!s->count || x > s->max)
        s->max = x;

    s->sum += x;
    s->sum2 += x*x;
    s->count += 1;
}


/*
 * stat_merge() implementation.
 */
void stat_merge(struct Stat *to, const struct Stat *from)
{
    if (!from->count)
        return; // nothing to merge

    if (!to->count || from->min < to->min)
        to->min = from->min;
    if (!to->count || from->max > to->max)
        to->max = from->max;
    to->sum += from->sum;
    to->sum2 += from->sum2;
    to->count += from->count;
}


/*
 * stat_print() implementation.
 */
void stat_print(const struct Stat *s, FILE *f)
{
    // TODO: more output formats!

    if (s->count)
    {
//        const double avg = s->sum/s->count;
//        const double var = s->sum2/s->count - avg*avg;
//        const double stdev = sqrt(var);
//        const double sigma = 2.0;

        fprintf(f, "{\"sum2\":%f, \"sum\":%f, \"min\":%f, \"max\":%f, \"count\":%llu}",
             s->sum2, s->sum, s->min, s->max, (long long unsigned int)s->count);
    }
    else
    {
        fprintf(f, "{\"sum2\":0, \"sum\":0, \"min\":null, \"max\":null, \"count\":0}");
    }
}
