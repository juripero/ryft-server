/** @file
 * @brief Main application.
 */
#include "conf.h"
#include "misc.h"
#include "json.h"

#include <signal.h>
#include <stddef.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>
#include <math.h>

#include <stdatomic.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/mman.h>
#include <unistd.h>
#include <fcntl.h>

#include <pthread.h>


// some global variables
static int volatile g_stopped = 0;

/**
 * @brief Handle system signals.
 * @param[in] signo The signal number.
 */
static void signal_handler(int signo)
{
    switch (signo)
    {
    case SIGINT:
        vlog1("SIGINT received, stopping...\n");
        g_stopped = 1; // stop main loop
        break;

    case SIGTERM:
        vlog1("SIGTERM received, stopping...\n");
        g_stopped = 1; // stop main loop
        break;

    default:
        vlog1("%d signal received, do nothing\n", signo);
        break;
    }
}


/**
 * @brief Parse the INDEX information.
 *
 * Tries to parse the INDEX line in the following format:
 * `filename,offset,length,fuzziness`.
 * The `filename`, `offset` and `fuzziness` are ignored.
 * So just `length` is parsed to the `data_len`.
 *
 * @param[in] idx_beg Begin of INDEX.
 * @param[in] idx_end End of INDEX.
 * @param[out] data_len Length of DATA in bytes.
 * @return Zero on success.
 */
static int parse_index(const uint8_t *idx_beg, const uint8_t *idx_end, uint64_t *data_len)
{
    const int COMMA = ',';

    // find ",fuzziness"
    const uint8_t *c4 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                idx_end - idx_beg);
    if (!c4)
        return -1; // no ",fuzziness" found

    // find ",length"
    const uint8_t *c3 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                c4 - idx_beg);
    if (!c3)
        return -2; // no ",length" found

    uint8_t *end = 0;
    *data_len = strtoull((const char*)c3+1, // +1 to skip comma
                         (char**)&end, 10);
    if (end != c4)
        return -3; // failed to parse length

    return 0; // OK
}


// TODO: abstract classes, dedicated files, etc
struct Stat {
    uint64_t count;
    double sum, sum2;
    double min, max;
};

// calculate statistics
static void stat_add(struct Stat *s, double x)
{
    if (!s->count || x < s->min)
        s->min = x;
    if (!s->count || x > s->max)
        s->max = x;
    s->sum += x;
    s->sum2 += x*x;
    s->count += 1;
}

// print statistics to STDOUT
static void stat_print(const struct Stat *s)
{
    if (s->count)
    {
        const double avg = s->sum/s->count;
//        const double var = s->sum2/s->count - avg*avg;
//        const double stdev = sqrt(var);
//        const double sigma = 2.0;

        vlog("{\"avg\":%f, \"sum\":%f, \"min\":%f, \"max\":%f, \"count\":%llu}\n",
             avg, s->sum, s->min, s->max, s->count);
    }
    else
    {
        vlog("{\"avg\":null, \"sum\":0, \"min\":null, \"max\":null, \"count\":0}\n");
    }
}

// clone statistics engine
static struct Stat* stat_clone(struct Stat *base)
{
    struct Stat *s = (struct Stat*)malloc(sizeof(*s));
    if (!s)
        return 0; // failed

    memcpy(s, base, sizeof(*s));
    return s;
}

// merge statistics
static void stat_merge(struct Stat *to, struct Stat *from)
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

// release statistics
static void stat_free(struct Stat *s)
{
    free(s);
}

// global instance, TODO: remove this
struct Stat g_stat;


/**
 * @brief Process DATA record.
 * @param[in] cfg Application configuration.
 * @param[in] dat Begin of DATA.
 * @param[in] len Length of DATA in bytes.
 * @return Zero on success.
 */
static int process_record(const struct Conf *cfg,
                          struct JSON_Field *root,
                          const uint8_t *beg,
                          const uint8_t *end,
                          struct Stat *s)
{
    (void)cfg; // not used yet

//    printf("  RECORD[%llu]:", len);
//    for (; len > 0; --len)
//        printf("%c", *dat++);
//    printf("\n");

    struct JSON_Field *field = root;
    while (field->children != 0)
        field = field->children;
    field->token.type = JSON_EOF;

    struct JSON_Parser parser;
    json_init(&parser, beg, end);

    if (!!json_get(&parser, root))
    {
        verr("ERROR: failed to get JSON field\n");
        return -1;
    }

    if (JSON_NUMBER != field->token.type)
    {
        verr("WARN: bad value found, ignored\n");
        return 0;
    }

    double x = strtod((const char*)field->token.beg, NULL);
    // vlog(" %g ", x);
    stat_add(s, x);

    return 0; // OK
}

// concurrency processing variables
struct {
    const struct Conf *cfg;
    struct JSON_Field *base_field;
    volatile int stopped;
    atomic_ptrdiff_t rpos; // read position, atomic
    atomic_ptrdiff_t wpos; // write position, atomic
    struct {
        const uint8_t *dat_beg;
        const uint8_t *dat_end;
    } buf[4*1024*1024];
} X;

// processing thread
static void* x_thread(void *p)
{
    struct Stat *s = (struct Stat*)p;
    struct JSON_Field *field = json_field_clone(X.base_field); // need to release at the end!
    // TODO: pass dedicated Task structure containing clone of Stat and clone of Field

    // TODO: count the number of processed records by this thread!
    int count = 0;

    while (1)
    {
        // get next processing index
        ptrdiff_t rpos = atomic_fetch_add(&X.rpos, 1);

        // busy-wait available data
        while (1)
        {
            ptrdiff_t wpos = atomic_load(&X.wpos);
            if (rpos < wpos)
                break; // finally have data to process
            if (X.stopped)
            {
                vlog1("processed records by x-thread: %d\n", count);
                return p; // done
            }
            pthread_yield(); // try again a bit later
        }

        const uint8_t *dat_beg = X.buf[rpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_beg;
        const uint8_t *dat_end = X.buf[rpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_end;
        int res = process_record(X.cfg, field, dat_beg, dat_end, s);
        if (res != 0)
        {
            verr("ERROR: failed to process RECORD: %s\n", res);// TODO: add "at" information from dat_p
            return 0; // failed
        }

        count += 1;
    }
}


/**
 * @brief Do the work.
 * @param[in] cfg Application configuration.
 * @param[in] field Head of JSON fields tree.
 * @param[in] idx_p The begin of INDEX file.
 * @param[in] idx_len The length of INDEX file in bytes.
 * @param[in] dat_p The begin of DATA file.
 * @param[in] dat_len The length of DATA file in bytes.
 * @return Zero on success.
 */
static int do_work(const struct Conf *cfg, struct JSON_Field *field,
                   const uint8_t *idx_beg, const uint8_t *idx_end,
                   const uint8_t *dat_beg, const uint8_t *dat_end)
{
    // remove DATA header
    if ((ptrdiff_t)cfg->header_len <= (dat_end - dat_beg))
        dat_beg += cfg->header_len;
    else
    {
        verr("ERROR: no DATA available (%d) to skip header (%d)\n",
             (dat_end - dat_beg), cfg->header_len);
        return -1; // failed
    }

    // remove DATA footer
    if ((off_t)cfg->footer_len <= (dat_end - dat_beg))
        dat_end -= cfg->footer_len;
    else
    {
        verr("ERROR: no DATA available (%d) to skip footer (%d)\n",
             dat_end - dat_beg, cfg->footer_len);
        return -1; // failed
    }

    // initialize concurrency stuff
    struct {
        pthread_t thread_id;
        struct Stat *stat;
    } *xx = 0;
    if (cfg->concurrency > 1)
    {
        vlog2("run with %d threads\n", cfg->concurrency);
        X.cfg = cfg;
        X.base_field = field;
        X.stopped = 0;
        X.rpos = 0;
        X.wpos = 0;

        xx = malloc(cfg->concurrency * sizeof(*xx));
        for (int i = 0; i < cfg->concurrency; ++i)
        {
            xx[i].stat = stat_clone(&g_stat);
            pthread_create(&xx[i].thread_id,
                           NULL/*attributes*/,
                           x_thread, xx[i].stat);
        }
    }

    // read INDEX line by line
    uint64_t count = 0;
    while ((idx_end - idx_beg) > 0)
    {
        // try to find the NEWLINE '\n' character
        const uint8_t *eol = (const uint8_t*)memchr(idx_beg, '\n',
                                                    idx_end - idx_beg);
        const uint8_t *next = eol ? (eol + 1) : idx_end;

        uint64_t d_len = 0;
        int res = parse_index(idx_beg, next, &d_len);
        if (res != 0)
        {
            verr("ERROR: failed to parse INDEX: %d\n", res); // TODO: add "at" information from idx_p
            return -2; // failed
        }

        // TODO: concurrency!!!
        if ((ptrdiff_t)d_len <= (dat_end - dat_beg))
        {
            if (cfg->concurrency > 1)
            {
                // put data to buffer for a processing thread
                ptrdiff_t wpos = atomic_load(&X.wpos);
                X.buf[wpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_beg = dat_beg;
                X.buf[wpos % (sizeof(X.buf)/sizeof(X.buf[0]))].dat_end = dat_beg + d_len;
                atomic_fetch_add(&X.wpos, 1);
            }
            else
            {
            int res = process_record(cfg, field, dat_beg,
                                     dat_beg + d_len,
                                     &g_stat);
            if (res != 0)
            {
                verr("ERROR: failed to process RECORD: %s\n", res);// TODO: add "at" information from dat_p
                return -3; // failed
            }
            }

            // go to next record...
            dat_beg += d_len + cfg->delim_len;
        }
        else
        {
            verr("ERROR: no DATA found\n"); // TODO: add "at" information from dat_p
            return -4; // failed
        }

        // go to next line
        idx_beg = next;
        count += 1;

        if (g_stopped != 0)
            break;
    }

    // merge concurrency engines if needed
    if (cfg->concurrency > 1)
    {
        X.stopped = 1; // stop all processing threads

        for (int i = 0; i < cfg->concurrency; ++i)
        {
            pthread_join(xx[i].thread_id, NULL);
            stat_merge(&g_stat, xx[i].stat);
            stat_free(xx[i].stat);
        }
        free(xx);
    }

    vlog2("total records processed: %llu\n", count);
    return 0; // OK
}


/**
 * @brief Application entry point.
 * @param[in] argc Number of command line arguments.
 * @param[in] argv List of command line arguments.
 * @return Zero on success.
 */
int main(int argc, const char *argv[])
{
    if (0)
    {
        extern void json_test(void);
        json_test();
        return 0; // OK
    }

    // setup signal handlers
    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);

    struct Conf cfg;
    memset(&cfg, 0, sizeof(cfg));
    if (!!conf_parse(&cfg, argc, argv))
        return -1;

    // print current configuration
    if (verbose >= 3)
        conf_print(&cfg);

    struct JSON_Field *field = 0;
    if (!!json_field_parse(&field, cfg.field))
    {
        verr("ERROR: failed to parse field \"%s\"\n", cfg.field);
        return -1;
    }

    // try to open INDEX file
    vlog2("opening INDEX file: %s\n", cfg.idx_path);
    int idx_fd = open(cfg.idx_path, O_RDONLY/*|O_LARGEFILE*/);
    if (idx_fd < 0)
    {
        verr("ERROR: failed to open INDEX file: %s\n",
             strerror(errno));
        return -1;
    }
    // and get INDEX file size
    struct stat idx_stat;
    memset(&idx_stat, 0, sizeof(idx_stat));
    if (!!fstat(idx_fd, &idx_stat))
    {
        verr("ERROR: failed to stat INDEX file: %s\n",
             strerror(errno));
        return -1;
    }
    vlog2("        INDEX file: #%d (%d bytes)\n",
          idx_fd, idx_stat.st_size);

    // try to open DATA file
    vlog2("opening  DATA file: %s\n", cfg.dat_path);
    int dat_fd = open(cfg.dat_path, O_RDONLY/*|O_LARGEFILE*/);
    if (dat_fd < 0)
    {
        verr("ERROR: failed to open DATA file: %s\n",
             strerror(errno));
        return -1;
    }
    // and get DATA file size
    struct stat dat_stat;
    memset(&dat_stat, 0, sizeof(dat_stat));
    if (!!fstat(dat_fd, &dat_stat))
    {
        verr("ERROR: failed to stat DATA file: %s\n",
             strerror(errno));
        return -1;
    }
    vlog2("         DATA file: #%d (%d bytes)\n",
          dat_fd, dat_stat.st_size);

    // TODO: do memory mapping part-by-part
    // let say 64MB per each part.

    // do memory mapping
    void *idx_p = mmap(0, idx_stat.st_size, PROT_READ, MAP_SHARED, idx_fd, 0);
    if (MAP_FAILED == idx_p)
    {
        verr("ERROR: failed to map INDEX file: %s\n",
             strerror(errno));
        return -1;
    }
    void *dat_p = mmap(0, dat_stat.st_size, PROT_READ, MAP_SHARED, dat_fd, 0);
    if (MAP_FAILED == dat_p)
    {
        verr("ERROR: failed to map DATA file: %s\n",
             strerror(errno));
        return -1;
    }

    // do actual processing
    do_work(&cfg, field,
            (const uint8_t*)idx_p, (const uint8_t*)idx_p + idx_stat.st_size,
            (const uint8_t*)dat_p, (const uint8_t*)dat_p + dat_stat.st_size);

    // print global statistics
    stat_print(&g_stat);

    // release resources
    munmap(idx_p, idx_stat.st_size);
    munmap(dat_p, dat_stat.st_size);
    close(dat_fd);
    close(idx_fd);
    conf_free(&cfg);

    return 0;
}

#include <stdarg.h>
#include "misc.h"

// print buffer
static void print_buf(const void *beg_, const void *end_)
{
    const char *beg = (const char*)beg_;
    const char *end = (const char*)end_;

    printf("<");
    while (beg != end)
        printf("%c", *beg++);
    printf(">");
}

// test token parsing
static int json_test_token(const char *json, ...)
{
    struct JSON_Parser p;
    json_init(&p, json, json + strlen(json));

    const int trace = 0;
    if (trace)
    {
        printf("parsing JSON:");
        print_buf(p.beg, p.end);
        printf("\n");
    }

    va_list args;
    va_start(args, json);

    while (1)
    {
        // parse token
        struct JSON_Token t;
        if (!!json_next(&p, &t))
        {
            printf("  FAILED\n");
            break;
        }

        if (t.type == JSON_EOF)
        {
            if (trace)
            printf("  #EOF\n");
            break;
        }

        if (trace)
        {
            printf("  #%d ", t.type);
            print_buf(t.beg, t.end);
            printf("\n");
        }

        const int expected_type = va_arg(args, int);
        const char *expected_token = va_arg(args, const char*);

        if ((int)t.type != expected_type)
        {
            verr("FAILED: bad token type: %d != %d\n", t.type, expected_type);
            return -1;
        }

        if (0 != memcmp(expected_token, t.beg, t.end - t.beg))
        {
            verr("FAILED: bad token\n");
            return -1;
        }
    }

    va_end(args);
    return 0; // OK
}

// test get by name
static int json_test_get_by_name1(const char *json, const char *name, int expected_type, const char *expected_token)
{
    struct JSON_Parser p;
    json_init(&p, json, json + strlen(json));

    const int trace = 0;
    if (trace)
    {
        printf("parsing JSON:");
        print_buf(p.beg, p.end);
        printf("\n");
    }

    struct JSON_Field field;
    memset(&field, 0, sizeof(field));
    strcpy(field.by_name, name);

    if (!!json_get(&p, &field))
    {
        printf("  FAILED\n");
    }

    struct JSON_Token *t = &field.token;

    if (trace)
    {
        printf("  #%d ", t->type);
        print_buf(t->beg, t->end);
        printf("\n");
    }

    if ((int)t->type != expected_type)
    {
        verr("FAILED: bad token type: %d != %d\n", t->type, expected_type);
        return -1;
    }

    if (0 != memcmp(expected_token, t->beg, t->end - t->beg))
    {
        verr("FAILED: bad token\n");
        return -1;
    }

    return 0; // OK
}


// test get by name
static int json_test_get_by_name11(const char *json, const char *name1, const char *name2, int expected_type, const char *expected_token)
{
    struct JSON_Parser p;
    json_init(&p, json, json + strlen(json));

    const int trace = 0;
    if (trace)
    {
        printf("parsing JSON:");
        print_buf(p.beg, p.end);
        printf("\n");
    }

    struct JSON_Field field;
    memset(&field, 0, sizeof(field));
    strcpy(field.by_name, name2);

    struct JSON_Field top_field;
    memset(&top_field, 0, sizeof(top_field));
    strcpy(top_field.by_name, name1);
    top_field.children = &field;

    if (!!json_get(&p, &top_field))
    {
        printf("  FAILED\n");
    }

    struct JSON_Token *t = &field.token;

    if (trace)
    {
        printf("  #%d ", t->type);
        print_buf(t->beg, t->end);
        printf("\n");
    }

    if ((int)t->type != expected_type)
    {
        verr("FAILED: bad token type: %d != %d\n", t->type, expected_type);
        return -1;
    }

    if (0 != memcmp(expected_token, t->beg, t->end - t->beg))
    {
        verr("FAILED: bad token\n");
        return -1;
    }

    return 0; // OK
}

// test get by name
static int json_test_get_by_name2(const char *json,
                                  const char *name1, int expected_type1, const char *expected_token1,
                                  const char *name2, int expected_type2, const char *expected_token2)
{
    struct JSON_Parser p;
    json_init(&p, json, json + strlen(json));

    const int trace = 0;
    if (trace)
    {
        printf("parsing JSON:");
        print_buf(p.beg, p.end);
        printf("\n");
    }

    struct JSON_Field field1;
    memset(&field1, 0, sizeof(field1));
    strcpy(field1.by_name, name1);

    struct JSON_Field field2;
    memset(&field2, 0, sizeof(field2));
    strcpy(field2.by_name, name2);

    field1.siblings = &field2;

    if (!!json_get(&p, &field1))
    {
        printf("  FAILED\n");
    }

    struct JSON_Token *t1 = &field1.token;
    struct JSON_Token *t2 = &field2.token;

    if (trace)
    {
        printf("  A #%d ", t1->type);
        print_buf(t1->beg, t1->end);
        printf("\n");
        printf("  B #%d ", t2->type);
        print_buf(t2->beg, t2->end);
        printf("\n");
    }

    if ((int)t1->type != expected_type1)
    {
        verr("FAILED: bad token type: %d != %d\n", t1->type, expected_type1);
        return -1;
    }

    if (0 != memcmp(expected_token1, t1->beg, t1->end - t1->beg))
    {
        verr("FAILED: bad token1\n");
        return -1;
    }

    if ((int)t2->type != expected_type2)
    {
        verr("FAILED: bad token type: %d != %d\n", t2->type, expected_type2);
        return -1;
    }

    if (0 != memcmp(expected_token2, t2->beg, t2->end - t2->beg))
    {
        verr("FAILED: bad token2\n");
        return -1;
    }

    return 0; // OK
}

// test field parsing
static int json_test_field(const char *path, int no_fields, ...)
{
    struct JSON_Field *f;
    if (!!json_field_parse(&f, path))
    {
        verr("FAILED: unable to parse field\n");
    }

    va_list args;
    va_start(args, no_fields);

    struct JSON_Field *ff = f;
    for (int i = 0; i < no_fields; ++i)
    {
        if (!ff)
        {
            verr("FAILED: no field\n");
            break;
        }
        const int expected_index = va_arg(args, int);
        if (ff->by_index != expected_index)
        {
            verr("FAILED: bad index\n");
            break;
        }
        const char *expected_name = va_arg(args, const char*);
        if (0 != strcmp(ff->by_name, expected_name))
        {
            verr("FAILED: bad name\n");
            break;
        }

        ff = ff->children;
    }

    json_field_free(f);

    va_end(args);
    return 0; // OK
}


/**
 * @brief Do the JSON parser tests.
 */
void json_test(void)
{
    const int all = 1;

    if (all)
    {
    json_test_token("");
    json_test_token("  false  \t", JSON_FALSE, "false");
    json_test_token("  \t  true  ", JSON_TRUE, "true");
    json_test_token("  \t  null \t\n\r ", JSON_NULL, "null");

    json_test_token(" { ", JSON_OBJECT_BEG, "{");
    json_test_token(" } ", JSON_OBJECT_END, "}");
    json_test_token(" [ ", JSON_ARRAY_BEG, "[");
    json_test_token(" ] ", JSON_ARRAY_END, "]");
    json_test_token(" : , ", JSON_COLON, ":", JSON_COMMA, ",");

    json_test_token(" 123  ", JSON_NUMBER, "123");
    json_test_token("123", JSON_NUMBER, "123");
    json_test_token("123.456", JSON_NUMBER, "123.456");
    json_test_token("123e10,", JSON_NUMBER, "123e10", JSON_COMMA, ",");

    json_test_token("\"\"", JSON_STRING, "");
    json_test_token(" \"a\"", JSON_STRING, "a");
    json_test_token(" \"b\" ", JSON_STRING, "b");
    json_test_token(" \"c\\n\" ", JSON_STRING_ESC, "c\\n");
    json_test_token(" \"d\\u1234\" ", JSON_STRING_ESC, "d\\u1234");
    json_test_token("\"key\":\"val\"", JSON_STRING, "key", JSON_COLON, ":", JSON_STRING, "val");
    }

    if (all)
    {
    json_test_get_by_name1("{}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":false}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":true}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":null}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":123}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":123.456}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":\"str\"}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":\"str\\u1234\"}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":[]}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":[0]}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":[0,1,2]}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":{}}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":{\"a\":\"b\"}}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":{\"a\":\"b\", \"c\":[{}]}}", "foo", JSON_EOF, "");
    json_test_get_by_name1("{\"test\":{\"a\":\"b\", \"c\":[0,1,2,3,[],[4,5],{\"a\":[]} ]}}", "foo", JSON_EOF, "");

    json_test_get_by_name1("{\"foo\":false}", "foo", JSON_FALSE, "false");
    json_test_get_by_name1("{\"foo\":true}", "foo", JSON_TRUE, "true");
    json_test_get_by_name1("{\"foo\":null}", "foo", JSON_NULL, "null");
    json_test_get_by_name1("{\"foo\":123}", "foo", JSON_NUMBER, "123");
    json_test_get_by_name1("{\"foo\":123.456}", "foo", JSON_NUMBER, "123.456");
    json_test_get_by_name1("{\"foo\":\"str\"}", "foo", JSON_STRING, "str");
    json_test_get_by_name1("{\"foo\":\"str\\u1234\"}", "foo", JSON_STRING_ESC, "str\\u1234");
    json_test_get_by_name1("{\"foo\":[]}", "foo", JSON_ARRAY, "[]");
    json_test_get_by_name1("{\"foo\":[0]}", "foo", JSON_ARRAY, "[0]");
    json_test_get_by_name1("{\"foo\":[0,1,2]}", "foo", JSON_ARRAY, "[0,1,2]");
    json_test_get_by_name1("{\"foo\":{}}", "foo", JSON_OBJECT, "{}");
    json_test_get_by_name1("{\"foo\":{\"a\":\"b\"}}", "foo", JSON_OBJECT, "{\"a\":\"b\"}");
    json_test_get_by_name1("{\"foo\":{\"a\":\"b\", \"c\":[{}]}}", "foo", JSON_OBJECT, "{\"a\":\"b\", \"c\":[{}]}");
    }

    if (all)
    {
    json_test_get_by_name11("{ \"foo\" : { \"no-bar\" : false } } ", "foo", "bar", JSON_EOF, "");
    json_test_get_by_name11("{ \"foo\" : { \"bar\" : false } } ", "foo", "bar", JSON_FALSE, "false");
    json_test_get_by_name11("{ \"foo\" : { \"bar\" : true } } ", "foo", "bar", JSON_TRUE, "true");
    json_test_get_by_name11("{ \"foo\" : { \"bar\" : null } } ", "foo", "bar", JSON_NULL, "null");
    json_test_get_by_name11("{ \"foo\" : { \"bar\" : 123 } } ", "foo", "bar", JSON_NUMBER, "123");
    json_test_get_by_name11("{ \"foo\" : { \"bar\" : 123.456 } } ", "foo", "bar", JSON_NUMBER, "123.456");
    json_test_get_by_name11("{ \"foo\" : { \"bar\" : [0,1,2] } } ", "foo", "bar", JSON_ARRAY, "[0,1,2]");
    }

    if (all)
    {
    json_test_get_by_name2("{ \"no-foo\" : false, \"no-bar\" : true } ", "foo", JSON_EOF, "", "bar", JSON_EOF, "");
    json_test_get_by_name2("{ \"foo\" : false, \"no-bar\" : true } ", "foo", JSON_FALSE, "false", "bar", JSON_EOF, "");
    json_test_get_by_name2("{ \"no-foo\" : false, \"bar\" : true } ", "foo", JSON_EOF, "", "bar", JSON_TRUE, "true");
    json_test_get_by_name2("{ \"foo\" : false, \"bar\" : true } ", "foo", JSON_FALSE, "false", "bar", JSON_TRUE, "true");
    json_test_get_by_name2("{ \"foo\" : {}, \"bar\" : [] } ", "foo", JSON_OBJECT, "{}", "bar", JSON_ARRAY, "[]");
    json_test_get_by_name2("{ \"foo\" : {\"a\" : []}, \"bar\" : [{},{}] } ", "foo", JSON_OBJECT, "{\"a\" : []}", "bar", JSON_ARRAY, "[{},{}]");
    }

    if (all)
    {
    json_test_field("foo", 1, -1, "foo");
    json_test_field(".foo", 1, -1, "foo");
    json_test_field("foo.", 1, -1, "foo");
    json_test_field(".foo.", 1, -1, "foo");
    json_test_field("..foo", 1, -1, "foo");
    json_test_field("foo..", 1, -1, "foo");
    json_test_field("..foo..", 1, -1, "foo");
    json_test_field("\"foo\"", 1, -1, "foo");
    json_test_field(".\"foo\"", 1, -1, "foo");
    json_test_field("\"foo\".", 1, -1, "foo");
    json_test_field(".\"foo\".", 1, -1, "foo");

    json_test_field("foo.bar", 2, -1, "foo", -1, "bar");
    json_test_field("foo.[1]", 2, -1, "foo", 0, "");
    json_test_field("foo.[101].[102]", 3, -1, "foo", 100, "", 101, "");
    }

    printf("OK\n");
}
