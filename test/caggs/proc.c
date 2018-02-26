#include "proc.h"
#include "misc.h"
#include "stat.h"
#include "json.h"

#include <string.h>
#include <stdlib.h>

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

    w->stat = stat_make();
    return w; // OK
}


/*
 * work_free() implementation.
 */
void work_free(struct Work *w)
{
    stat_free(w->stat);
    json_field_free(w->field);
    free(w);
}


/*
 * work_do() implementation.
 */
int work_do(struct Work *w, const uint8_t *data_buf,
            const struct RecordRef *records,
            uint64_t num_of_records)
{
    // process all records
    for (uint64_t i = 0; i < num_of_records; ++i)
    {
        // printf("record #%"PRIu64" at %"PRIu64" of %"PRIu64" bytes: ",
        //        i, records[i].offset, records[i].length);

        const uint8_t *rec = data_buf + records[i].offset;
        const uint8_t *end = rec + records[i].length;
        // extern void print_buf(const void*, const void*);
        // print_buf(rec, end); printf("\n");

        struct JSON_Field *field = w->field;
        while (field->children != 0)
            field = field->children;
        field->token.type = JSON_EOF;

        struct JSON_Parser parser;
        json_init(&parser, rec, end);

        if (!!json_get(&parser, w->field))
        {
            verr1("ERROR: failed to get JSON field\n");
            // return -1; // failed
            continue;
        }

        if (JSON_NUMBER != field->token.type)
        {
            w->stat->count += 1; // TODO: check not NULL
            verr2("WARN: bad value found, ignored\n");
            return 0;
        }

        double x = strtod((const char*)field->token.beg, NULL);
        // vlog(" %g ", x);
        stat_add(w->stat, x);
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



//// concurrency processing variables
//struct {
//    const struct Conf *cfg;
//    struct JSON_Field *base_field;
//    volatile int stopped;
//    atomic_ptrdiff_t rpos; // read position, atomic
//    atomic_ptrdiff_t wpos; // write position, atomic
//    struct {
//        const uint8_t *dat_beg;
//        const uint8_t *dat_end;
//    } buf[4*1024*1024];
//} X;

//// processing thread
//static void* x_thread(void *p)
//{
//    struct Stat *s = (struct Stat*)p;
//    struct JSON_Field *field = json_field_clone(X.base_field); // need to release at the end!
//    // TODO: pass dedicated Task structure containing clone of Stat and clone of Field

//    // TODO: count the number of processed records by this thread!
//    int count = 0;

//    while (1)
//    {
//        // get next processing index
//        ptrdiff_t rpos = atomic_fetch_add(&X.rpos, 1);

//        // busy-wait available data
//        while (1)
//        {
//            ptrdiff_t wpos = atomic_load(&X.wpos);
//            if (rpos < wpos)
//                break; // finally have data to process
//            if (X.stopped)
//            {
//                vlog1("processed records by x-thread: %d\n", count);
//                return p; // done
//            }
//            pthread_yield(); // try again a bit later
//        }

//        const uint8_t *dat_beg = X.buf[rpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_beg;
//        const uint8_t *dat_end = X.buf[rpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_end;
//        int res = process_record(X.cfg, field, dat_beg, dat_end, s);
//        if (res != 0)
//        {
//            verr("ERROR: failed to process RECORD: %s\n", res);// TODO: add "at" information from dat_p
//            return 0; // failed
//        }

//        count += 1;
//    }
//}


///**
// * @brief Do the work.
// * @param[in] cfg Application configuration.
// * @param[in] field Head of JSON fields tree.
// * @param[in] idx_p The begin of INDEX file.
// * @param[in] idx_len The length of INDEX file in bytes.
// * @param[in] dat_p The begin of DATA file.
// * @param[in] dat_len The length of DATA file in bytes.
// * @return Zero on success.
// */
//static int do_work0(const struct Conf *cfg, struct JSON_Field *field,
//                   const uint8_t *idx_beg, const uint8_t *idx_end,
//                   const uint8_t *dat_beg, const uint8_t *dat_end)
//{
//    // remove DATA header
//    if ((ptrdiff_t)cfg->header_len <= (dat_end - dat_beg))
//        dat_beg += cfg->header_len;
//    else
//    {
//        verr("ERROR: no DATA available (%d) to skip header (%d)\n",
//             (dat_end - dat_beg), cfg->header_len);
//        return -1; // failed
//    }

//    // remove DATA footer
//    if ((off_t)cfg->footer_len <= (dat_end - dat_beg))
//        dat_end -= cfg->footer_len;
//    else
//    {
//        verr("ERROR: no DATA available (%d) to skip footer (%d)\n",
//             dat_end - dat_beg, cfg->footer_len);
//        return -1; // failed
//    }

//    // initialize concurrency stuff
//    struct {
//        pthread_t thread_id;
//        struct Stat *stat;
//    } *xx = 0;
//    if (cfg->concurrency > 1)
//    {
//        vlog2("run with %d threads\n", cfg->concurrency);
//        X.cfg = cfg;
//        X.base_field = field;
//        X.stopped = 0;
//        X.rpos = 0;
//        X.wpos = 0;

//        xx = malloc(cfg->concurrency * sizeof(*xx));
//        for (int i = 0; i < cfg->concurrency; ++i)
//        {
//            xx[i].stat = stat_clone(&g_stat);
//            pthread_create(&xx[i].thread_id,
//                           NULL/*attributes*/,
//                           x_thread, xx[i].stat);
//        }
//    }

//    // read INDEX line by line
//    uint64_t count = 0;
//    while ((idx_end - idx_beg) > 0)
//    {
//        // try to find the NEWLINE '\n' character
//        const uint8_t *eol = (const uint8_t*)memchr(idx_beg, '\n',
//                                                    idx_end - idx_beg);
//        const uint8_t *next = eol ? (eol + 1) : idx_end;

//        uint64_t d_len = 0;
//        int res = parse_index(idx_beg, next, &d_len);
//        if (res != 0)
//        {
//            verr("ERROR: failed to parse INDEX: %d\n", res); // TODO: add "at" information from idx_p
//            return -2; // failed
//        }

//        // TODO: concurrency!!!
//        if ((ptrdiff_t)d_len <= (dat_end - dat_beg))
//        {
//            if (cfg->concurrency > 1)
//            {
//                // put data to buffer for a processing thread
//                ptrdiff_t wpos = atomic_load(&X.wpos);
//                X.buf[wpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_beg = dat_beg;
//                X.buf[wpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_end = dat_beg + d_len;
//                atomic_fetch_add(&X.wpos, 1);
//            }
//            else
//            {
//            int res = process_record(cfg, field, dat_beg,
//                                     dat_beg + d_len,
//                                     &g_stat);
//            if (res != 0)
//            {
//                verr("ERROR: failed to process RECORD: %s\n", res);// TODO: add "at" information from dat_p
//                return -3; // failed
//            }
//            }

//            // go to next record...
//            dat_beg += d_len + cfg->delim_len;
//        }
//        else
//        {
//            verr("ERROR: no DATA found\n"); // TODO: add "at" information from dat_p
//            return -4; // failed
//        }

//        // go to next line
//        idx_beg = next;
//        count += 1;

//        if (g_stopped != 0)
//            break;
//    }

//    // merge concurrency engines if needed
//    if (cfg->concurrency > 1)
//    {
//        X.stopped = 1; // stop all processing threads

//        for (int i = 0; i < cfg->concurrency; ++i)
//        {
//            pthread_join(xx[i].thread_id, NULL);
//            stat_merge(&g_stat, xx[i].stat);
//            stat_free(xx[i].stat);
//        }
//        free(xx);
//    }

//    vlog2("total records processed: %"PRIu64"\n", count);
//    return 0; // OK
//}
