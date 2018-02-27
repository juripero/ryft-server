#include "proc.h"
#include "misc.h"
#include "stat.h"
#include "json.h"

#include <inttypes.h>
#include <string.h>
#include <stdlib.h>
#include <errno.h>

#include <pthread.h>

/*
 * parse_index() implementation.
 */
int parse_index(const uint8_t *idx_beg,
                const uint8_t *idx_end,
                uint64_t *data_len)
{
    const int COMMA = ',';

    //extern void print_buf(const void*, const void*);
    //printf("parsing INDEX: "); print_buf(idx_beg, idx_end-1); printf("\n");

    // find ",fuzziness"
    const uint8_t *c4 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                idx_end - idx_beg);
    if (!c4)
    {
        verr3("parse_index: no \",fuzziness\" found\n");
        return -1; // no ",fuzziness" found
    }

    // find ",length"
    const uint8_t *c3 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                c4 - idx_beg);
    if (!c3)
    {
        verr3("parse_index: no \",length\" found\n");
        return -2; // no ",length" found
    }

    uint8_t *end = 0;
    *data_len = strtoull((const char*)c3+1, // +1 to skip comma
                         (char**)&end, 10);
    if (end != c4)
    {
        verr3("parse_index: failed to parse length\n");
        return -3; // failed to parse length
    }

    return 0; // OK
}


/*
 * parse_index_chunk() implementation.
 */
int parse_index_chunk(int is_last_chunk,
                      const uint8_t *buf, uint64_t *len_,
                      uint64_t delim_len,
                      uint64_t data_start,
                      uint64_t *max_data_len_,
                      struct RecordRef *records,
                      uint64_t *num_of_records_)
{
    int64_t len = *len_; // INDEX buffer remain in bytes

    // total DATA length in bytes
    const uint64_t max_data_len = *max_data_len_;
    uint64_t data_len = 0;

    // number of indices/record references parsed
    const uint64_t max_count = *num_of_records_;
    uint64_t count = 0;

    int res = 0;

    // read INDEX line by line
    while (0 < len && count < max_count)
    {
        // try to find the NEWLINE '\n' character
        const uint8_t *eol = (const uint8_t*)memchr(buf, '\n', len);
        if (!eol && !is_last_chunk)
        {
            res = 1;
            break; // done, leave the last INDEX (part of it) to the next chunk
        }

        const uint8_t *next = eol ? (eol+1) : (buf+len);

        uint64_t d_len = 0;
        if (!!parse_index(buf, next, &d_len))
        {
            verr1("ERROR: failed to parse INDEX\n"); // TODO: add "at" information from idx_beg
            return -2; // failed
        }

        if (data_len + d_len + delim_len > max_data_len)
            break; // no space to save data

        records[count].offset = data_start + data_len;
        records[count].length = d_len;
        count += 1;

        // go to next record...
        data_len += d_len + delim_len;
        len -= (next - buf);
        buf = next;
    }

    // results
    *max_data_len_ = data_len;
    *num_of_records_ = count;
    *len_ -= len;

    return res; // OK
}


/**
 * @brief The single thread processing.
 */
struct XProc
{
    pthread_t thread_id;

    struct JSON_Field *field;   ///< @brief Field to search for.
    struct Stat *stat;          ///< @brief Intermediate statistics.

    const uint8_t *data_buf;            ///< @brief Begin of DATA buffer.
    const struct RecordRef *records;    ///< @brief Records to process.
    uint64_t num_of_records;            ///< @brief Number of records to process.
};


/*
 * work_make() implementation.
 */
struct Work* work_make(const struct Conf *cfg)
{
    struct Work *w = (struct Work*)malloc(sizeof(*w));
    if (!w)
        return w; // out of memory

    // parse field
    w->field = 0;
    if (!!json_field_parse(&w->field, cfg->field))
    {
        verr("ERROR: failed to parse field \"%s\"\n", cfg->field);
        free(w);
        return 0;
    }

    // final statistics
    w->stat = stat_make();

    // concurrency
    w->xproc_started = 0;
    if (cfg->concurrency > 0)
    {
        const int n = cfg->concurrency;

        struct XProc *xproc = (struct XProc*)malloc(n * sizeof(*xproc));
        for (int i = 0; i < n; ++i)
        {
            xproc[i].field = json_field_clone(w->field);
            xproc[i].stat = stat_clone(w->stat);
        }

        w->n_xproc = n;
        w->xproc = xproc;
    }
    else
    {
        w->n_xproc = 0;
        w->xproc = 0;
    }

    return w; // OK
}


/*
 * work_free() implementation.
 */
void work_free(struct Work *w)
{
    if (w->n_xproc > 0)
    {
        // release XProc processing units
        struct XProc *xproc = (struct XProc*)w->xproc;
        for (int i = 0; i < w->n_xproc; ++i)
        {
            json_field_free(xproc[i].field);
            stat_free(xproc[i].stat);
        }
    }

    stat_free(w->stat);
    json_field_free(w->field);
    free(w);
}


/**
 * @brief Processing thread.
 * @param param The XProc structure.
 */
static void* xproc_thread(void *param)
{
    struct XProc *x = (struct XProc*)param;

    // process all records
    extern int volatile g_stopped;
    for (uint64_t i = 0; !g_stopped && i < x->num_of_records; ++i)
    {
        // printf("record #%"PRIu64" at %"PRIu64" of %"PRIu64" bytes: ",
        //        i, records[i].offset, records[i].length);

        const uint8_t *rec = x->data_buf + x->records[i].offset;
        const uint8_t *end = rec + x->records[i].length;
        // extern void print_buf(const void*, const void*);
        // print_buf(rec, end); printf("\n");

        // TODO: need to reset all fields: children and siblings
        struct JSON_Field *field = x->field;
        while (field->children != 0)
            field = field->children;
        field->token.type = JSON_EOF;

        struct JSON_Parser parser;
        json_init(&parser, rec, end);

        if (!!json_get(&parser, x->field))
        {
            verr1("ERROR: failed to get JSON field\n");
            // return -1; // failed
            continue;
        }

        if (JSON_NUMBER != field->token.type)
        {
            x->stat->count += 1; // TODO: check not NULL
            verr2("WARN: bad value found, ignored\n");
            return 0;
        }

        double val = strtod((const char*)field->token.beg, NULL);
        // vlog(" %g ", val);
        stat_add(x->stat, val);
    }

    return param;
}


/*
 * work_do_start() implementation.
 */
int work_do_start(struct Work *w, const uint8_t *data_buf,
                  const struct RecordRef *records,
                  uint64_t num_of_records)
{
    const int n = w->n_xproc;

    if (!n)
    {
        struct XProc x;
        x.field = w->field;
        x.stat = w->stat;
        x.data_buf = data_buf;
        x.records = records;
        x.num_of_records = num_of_records;

        xproc_thread(&x); // do processing on the same thread!
        return 0; // OK
    }

    // start processing units (new iteration)
    const uint64_t N = (num_of_records + n-1) / n; // ceil
    vlog2("xproc: start %d threads (about %"PRId64" records per thread)\n", n, N);
    for (int i = 0; i < n; ++i)
    {
        w->xproc[i].data_buf = data_buf;
        w->xproc[i].records = records + N*i;
        w->xproc[i].num_of_records = (i+1 == n) ? (num_of_records - N*i) : N;
        stat_init(w->xproc[i].stat); // reset stat
        if (!!pthread_create(&w->xproc[i].thread_id,
                             NULL/*attributes*/,
                             xproc_thread,
                             &w->xproc[i]))
        {
            verr1("failed to create processing thread: %s\n",
                  strerror(errno));
            return -1;
        }
    }

    w->xproc_started = 1;
    return 0; // OK
}


/*
 * work_do_join() implementation.
 */
int work_do_join(struct Work *w)
{
    // wait all processing units (previous iteration)
    if (w->xproc_started)
    {
        const int n = w->n_xproc;
        vlog2("xproc: wait %d threads\n", n);
        for (int i = 0; i < n; ++i)
        {
            if (!!pthread_join(w->xproc[i].thread_id, NULL))
            {
                verr1("failed to join processing thread: %s\n",
                      strerror(errno));
                return -1;
            }

            // merge statistics
            stat_merge(w->stat, w->xproc[i].stat);
        }

        w->xproc_started = 0;
    }

    return 0; // OK
}


/*
 * work_print() implementation.
 */
void work_print(struct Work *w, FILE *f)
{
    stat_print(w->stat, f);
}
