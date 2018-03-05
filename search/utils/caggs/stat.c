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
        const double avg = s->sum/s->count;
//        const double var = s->sum2/s->count - avg*avg;
//        const double stdev = sqrt(var);
//        const double sigma = 2.0;

        fprintf(f, "{\"avg\":%f, \"sum\":%f, \"min\":%f, \"max\":%f, \"count\":%llu}",
             avg, s->sum, s->min, s->max, (long long unsigned int)s->count);
    }
    else
    {
        fprintf(f, "{\"avg\":null, \"sum\":0, \"min\":null, \"max\":null, \"count\":0}");
    }
}
