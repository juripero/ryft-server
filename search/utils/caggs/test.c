#include "misc.h"
#include "json.h"

#include <stdarg.h>
#include <string.h>

// print buffer
void print_buf(const void *beg_, const void *end_)
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

#ifndef NO_TESTS

/**
 * @brief Test entry point.
 * @param[in] argc Number of command line arguments.
 * @param[in] argv List of command line arguments.
 * @return Zero on success.
 */
int main(int argc, const char *argv[])
{
    (void)argc;
    (void)argv;
    json_test();
    return 0; // OK
}

#endif // NO_TESTS


/*
To test large file processing the following Bash script can be used:

./caggs -i /ryftone/test-100K.txt -d /ryftone/test-100K.bin --field foo -D2 -X1 --data-chunk=2G --index-chunk=2G -q > ref.log

for d in 1048576 1052676; do
    for i in {1048576..1052676}; do
        echo -n "data:$d index:$i: ... " >> my.log && \
        ./caggs -i /ryftone/test-100K.txt -d /ryftone/test-100K.bin --field foo \
            -D2 -X1 --data-chunk=$d --index-chunk=$i -q > tst.log && \
        diff ref.log tst.log && echo "OK" >> my.log || echo "FAILED" >> my.log
    done
done
*/
