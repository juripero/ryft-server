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

    struct JSON_Field *field_tree;  ///< @brief Field tree to search for.
    struct JSON_Field **fields;     ///< @brief Leaf fields.
    struct Stat **stats;            ///< @brief Intermediate statistics.
    int n_fields;                   ///< @brief The number of fields.

    const uint8_t *data_buf;            ///< @brief Begin of DATA buffer.
    const struct RecordRef *records;    ///< @brief Records to process.
    uint64_t num_of_records;            ///< @brief Number of records to process.
};


/**
 * @brief Get the last field from field tree.
 * @param field The field tree.
 * @return The last leaf field.
 */
static struct JSON_Field* field_get_last(struct JSON_Field *field)
{
    while (field->children != 0)
        field = field->children;
    return field;
}


// print JSON field
static void field_print(int ident, struct JSON_Field *field, FILE *f)
{
    while (field != 0)
    {
        for (int i = 0; i < ident; ++i)
            fprintf(f, "  ");
        if (field->by_index < 0)
            fprintf(f, "%s\n", field->by_name);
        else
            fprintf(f, "[%d]\n", field->by_index+JSON_INDEX_BASE);

        field_print(ident+1, field->children, f);
        field = field->siblings;
    }
}


/**
 * @brief Merge the field to the field tree.
 * @param root The field tree.
 * @param field The field to merge.
 * @return The last lead field.
 */
static struct JSON_Field* field_merge(struct JSON_Field *root, struct JSON_Field *field)
{
    if (!field)
        return 0;

    struct JSON_Field *sf = 0; // corresponding field in the 'root' tree
    if (field->by_index < 0)
    {
        // try to find by name
        const uint8_t* name_beg = (const uint8_t*)&field->by_name[0];
        const uint8_t* name_end = name_beg + strlen(field->by_name);
        sf = json_field_by_name(root, name_beg, name_end);
    }
    else
    {
        // try to find by index
        sf = json_field_by_index(root, field->by_index);
    }

    if (sf)
    {
        // corresponding field already exists
        struct JSON_Field *child = field->children;
        field->children = 0;    // prevent deleting rest of tree
        json_field_free(field); // delete existing node
        if (!child)
            return root;

        if (root == sf)
            sf = root->children;

        if (sf)
            return field_merge(sf, child);
        else
        {
            root->children = child;
            return field_get_last(child);
        }
    }
    else
    {
        // corresponding field is not found, add the rest of tree
        field->siblings = root->siblings;
        root->siblings = field;
        return field_get_last(field);
    }
}


/*
 * work_make() implementation.
 */
struct Work* work_make(const struct Conf *cfg)
{
    struct Work *w = (struct Work*)malloc(sizeof(*w));
    if (!w)
        return w; // out of memory

    // concurrency
    w->xproc_started = 0;
    if (cfg->concurrency > 0)
    {
        w->n_xproc = cfg->concurrency;
        w->xproc = (struct XProc*)malloc(w->n_xproc * sizeof(w->xproc[0]));
        for (int k = 0; k < w->n_xproc; ++k)
        {
            w->xproc[k].field_tree = 0;
            w->xproc[k].fields = (struct JSON_Field**)malloc(cfg->n_fields*sizeof(w->xproc[k].fields[0]));
            w->xproc[k].stats = (struct Stat**)malloc(cfg->n_fields*sizeof(w->xproc[k].stats[0]));
            w->xproc[k].n_fields = cfg->n_fields;
        }
    }
    else
    {
        w->n_xproc = 0;
        w->xproc = 0;
    }

    // parse fields
    w->field_tree = 0;
    w->fields = (struct JSON_Field**)malloc(cfg->n_fields*sizeof(w->fields[0]));
    w->stats = (struct Stat**)malloc(cfg->n_fields*sizeof(w->stats[0]));
    w->n_fields = cfg->n_fields;
    for (int i = 0; i < cfg->n_fields; ++i)
    {
        struct JSON_Field *field = 0;
        if (!!json_field_parse(&field, cfg->fields[i]))
        {
            verr("ERROR: failed to parse field \"%s\"\n", cfg->fields[i]);
            free(w);
            return 0;
        }

        struct JSON_Field *cc_field = w->n_xproc ? json_field_clone(field) : 0;
        if (!w->field_tree)
        {
            // first field use "as is"
            w->fields[i] = field_get_last(field);
            w->field_tree = field;
        }
        else
        {
            // merge the others
            w->fields[i] = field_merge(w->field_tree, field);
        }
        w->stats[i] = stat_make();

        // concurrency
        for (int k = 0; k < w->n_xproc; ++k)
        {
            field = json_field_clone(cc_field);
            if (!w->xproc[k].field_tree)
            {
                // first field use "as is"
                w->xproc[k].fields[i] = field_get_last(field);
                w->xproc[k].field_tree = field;
            }
            else
            {
                // merge the others
                w->xproc[k].fields[i] = field_merge(w->xproc[k].field_tree, field);
            }

            w->xproc[k].stats[i] = stat_clone(w->stats[i]);
        }

        json_field_free(cc_field);
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
        for (int k = 0; k < w->n_xproc; ++k)
        {
            struct XProc *x = &xproc[k];
            for (int i = 0; i < x->n_fields; ++i)
                stat_free(x->stats[i]);
            json_field_free(x->field_tree);
            free(x->fields);
            free(x->stats);
        }
    }

    for (int i = 0; i < w->n_fields; ++i)
        stat_free(w->stats[i]);
    json_field_free(w->field_tree);
    free(w->fields);
    free(w->stats);

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
    for (uint64_t k = 0; !g_stopped && k < x->num_of_records; ++k)
    {
        // printf("record #%"PRIu64" at %"PRIu64" of %"PRIu64" bytes: ",
        //        k, records[k].offset, records[k].length);

        const uint8_t *rec = x->data_buf + x->records[k].offset;
        const uint8_t *end = rec + x->records[k].length;
        // extern void print_buf(const void*, const void*);
        // print_buf(rec, end); printf("\n");

        // reset all fields
        for (int i = 0; i < x->n_fields; ++i)
            x->fields[i]->token.type = JSON_EOF;

        struct JSON_Parser parser;
        json_init(&parser, rec, end);

        if (!!json_get(&parser, x->field_tree))
        {
            verr1("ERROR: failed to get JSON field\n");
            // return -1; // failed
            continue;
        }

        for (int i = 0; i < x->n_fields; ++i)
        {
            struct JSON_Field *field = x->fields[i];
            struct Stat *stat = x->stats[i];

            if (JSON_NUMBER != field->token.type)
            {
                stat->count += 1; // TODO: check not NULL
                verr2("WARN: bad value found, ignored\n");
                return 0;
            }

            double val = strtod((const char*)field->token.beg, NULL);
            // vlog(" %g ", val);
            stat_add(stat, val);
        }
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
    w->xproc_start = get_time();

    if (!n)
    {
        struct XProc x;
        x.field_tree = w->field_tree;
        x.fields = w->fields;
        x.stats = w->stats;
        x.n_fields = w->n_fields;
        x.data_buf = data_buf;
        x.records = records;
        x.num_of_records = num_of_records;

        xproc_thread(&x); // do processing on the same thread!

        vlog3("xproc: done in %.3fms\n", (get_time() - w->xproc_start)*1e-3);
        return 0; // OK
    }

    // start processing units (new iteration)
    const uint64_t N = (num_of_records + n-1) / n; // ceil
    vlog2("xproc: start %d threads (about %"PRId64" records per thread)\n", n, N);
    for (int k = 0; k < n; ++k)
    {
        w->xproc[k].data_buf = data_buf;
        w->xproc[k].records = records + N*k;
        w->xproc[k].num_of_records = (k+1 == n) ? (num_of_records - N*k) : N;
        for (int i = 0; i < w->n_fields; ++i)
            stat_init(w->xproc[k].stats[i]); // reset stat
        if (!!pthread_create(&w->xproc[k].thread_id,
                             NULL/*attributes*/,
                             xproc_thread,
                             &w->xproc[k]))
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
        for (int k = 0; k < n; ++k)
        {
            if (!!pthread_join(w->xproc[k].thread_id, NULL))
            {
                verr1("failed to join processing thread: %s\n",
                      strerror(errno));
                return -1;
            }

            // merge statistics
            for (int i = 0; i < w->n_fields; ++i)
                stat_merge(w->stats[i], w->xproc[k].stats[i]);
        }

        vlog3("xproc: done in %.3fms\n", (get_time() - w->xproc_start)*1e-3);
        w->xproc_started = 0;
    }

    return 0; // OK
}


/*
 * work_print() implementation.
 */
void work_print(struct Work *w, FILE *f)
{
    fprintf(f, "[");
    for (int i = 0; i < w->n_fields; ++i)
    {
        if (i) fprintf(f, "\n,");
        stat_print(w->stats[i], f);
    }
    fprintf(f, "]");
}


// print work tree
void work_test(struct Work *w, FILE *f)
{
    fprintf(f, "final tree:\n");
    field_print(1, w->field_tree, f);
    for (int i = 0; i < w->n_fields; ++i)
    {
        fprintf(f, "field #%d:\n", i);
        field_print(1, w->fields[i], f);
    }
}
