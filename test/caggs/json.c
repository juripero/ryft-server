#include "json.h"

#include <string.h>
#include <stdlib.h>


/**
 * @brief The JSON parser's state.
 */
enum JSON_ParserState
{
    JSON_PARSER_STATE_VALUE,
    JSON_PARSER_STATE_OBJ_KEY,
    JSON_PARSER_STATE_OBJ_VAL,

    JSON_PARSER_STATE_EOF,
};

/*
 * json_init() implementation.
 */
void json_init(struct JSON_Parser *parser,
               const void *json_beg,
               const void *json_end)
{
    parser->beg = (const uint8_t*)json_beg;
    parser->end = (const uint8_t*)json_end;
    parser->state = JSON_PARSER_STATE_VALUE;
    parser->no_tokens = 0;
}


/**
 * @brief Skip all whitespaces.
 * @param[in] parser JSON parser.
 * @return Zero on success.
 *   Non-zero if parser is empty.
 */
static int json_skip_ws(struct JSON_Parser *parser)
{
    // iterate over JSON data
    while ((parser->end - parser->beg) > 0)
    {
        switch (parser->beg[0])
        {
            // ignore all whitespaces
            case '\t': case '\r':
            case '\n': case ' ':
                parser->beg += 1;
                break;

            // stop on first non-whitespace
            default:
                return 1;
        }
    }

    return 0; // no more data
}

static int json_skip_object(struct JSON_Parser *parser);
static int json_skip_array(struct JSON_Parser *parser);

/**
 * @brief Skip the OBJECT.
 *
 * The JSON_OBJECT_BEG should be already skipped.
 *
 * @param parser JSON parser.
 * @return Zero on success.
 */
static int json_skip_object(struct JSON_Parser *parser)
{
    while (1)
    {
        // we expects string as a key
        struct JSON_Token token;
        if (!!json_next(parser, &token))
            return -1; // failed
        if (JSON_OBJECT_END == token.type)
            break; // done
        if (JSON_STRING != token.type  && JSON_STRING_ESC != token.type)
            return -1; // failed, string key expected

        // we expects string as a key
        if (!!json_next(parser, &token))
            return -1; // failed
        if (JSON_COLON != token.type)
            return -1; // failed, colon expected

        // get value
        if (!!json_next(parser, &token))
            return -1; // failed
        switch (token.type)
        {
            case JSON_OBJECT_BEG:
                if (!!json_skip_object(parser))
                    return -1; // failed
                break;

            case JSON_ARRAY_BEG:
                if (!!json_skip_array(parser))
                    return -1; // failed
                break;

            // primitive
            case JSON_STRING:
            case JSON_STRING_ESC:
            case JSON_NUMBER:
            case JSON_FALSE:
            case JSON_TRUE:
            case JSON_NULL:
                // just ignored
                break;

            default:
                return -1; // failed
        }

        // we expects comma between elements
        if (!!json_next(parser, &token))
            return -1; // failed
        if (JSON_OBJECT_END == token.type)
            break; // done
        if (JSON_COMMA != token.type)
            return -1; // failed, comma expected
    }

    return 0; // OK
}


/**
 * @brief Skip the ARRAY.
 *
 * The JSON_ARRAY_BEG should be already skipped.
 *
 * @param parser JSON parser.
 * @return Zero on success.
 */
static int json_skip_array(struct JSON_Parser *parser)
{
    while (1)
    {
        // get array element
        struct JSON_Token token;
        if (!!json_next(parser, &token))
            return -1; // failed
        if (JSON_ARRAY_END == token.type)
            break; // done

        switch (token.type)
        {
            case JSON_OBJECT_BEG:
                if (!!json_skip_object(parser))
                    return -1; // failed
                break;

            case JSON_ARRAY_BEG:
                if (!!json_skip_array(parser))
                    return -1; // failed
                break;

            // primitive
            case JSON_STRING:
            case JSON_STRING_ESC:
            case JSON_NUMBER:
            case JSON_FALSE:
            case JSON_TRUE:
            case JSON_NULL:
                // just ignored
                break;

            default:
                return -1; // failed
        }

        // we expects comma between elements
        if (!!json_next(parser, &token))
            return -1; // failed
        if (JSON_ARRAY_END == token.type)
            break; // done
        if (JSON_COMMA != token.type)
            return -1; // failed, comma expected
    }

    return 0; // OK
}


/**
 * @brief Check the HEX character.
 * @param ch The character to check.
 * @return Non-zero on HEX: `0-9A-Fa-f`.
 */
static inline int json_is_hex(int ch)
{
    if (ch >= '0' && ch <= '9')
        return 1;
    if (ch >= 'A' && ch <= 'F')
        return 1;
    if (ch >= 'a' && ch <= 'f')
        return 1;

    return 0; // not a HEX
}


/**
 * @brief Get next string.
 * @param parser[in] JSON parser.
 * @param token[out] The JSON string token.
 * @return Zero on success.
 */
static inline int json_next_string(struct JSON_Parser *parser,
                                   struct JSON_Token *token)
{
    const uint8_t *p = parser->beg+1; // skip starting quote

    int escaped = 0;
    while ((parser->end - p) > 0)
    {
        // end of string
        if (*p == '\"')
        {
            ++p; // skip ending quote
            break;
        }

        // quoted symbol
        else if (*p == '\\')
        {
            ++p; // skip backslash
            if ((parser->end - p) <= 0)
            {
                // no data to get escaped symbol
                return -1; // failed
            }

            const int ch = *p;
            switch (ch)
            {
                // special escaped symbols
                case '\"': case '/': case 'b': case 'f':
                case '\\': case 'r': case 'n': case 't':
                    escaped = 1;
                    break;

                // \uXXXX escaped symbol
                case 'u':
                    ++p; // skip 'u'
                    if ((parser->end - p) < 4)
                    {
                        // no data to get escaped symbol
                        return -1; // failed
                    }
                    for (int i = 0; i < 4; ++i)
                        if (!json_is_hex(*p++))
                        {
                            // bad escaped symbol
                            return -1; // failed
                        }
                    escaped = 1;
                    break;

                default:
                    // bad escaped symbol
                    return -1; // failed
            }
        }

        // use "as is"
        else
            ++p;
    }

    token->beg = parser->beg+1; // without quote
    token->end = p-1;           // without quote
    token->type = escaped ? JSON_STRING_ESC
                          : JSON_STRING;
    parser->beg = p;
    return 0; // OK
}


/**
 * @brief Get next number.
 * @param parser[in] JSON parser.
 * @param token[out] The JSON number token.
 * @return Zero on success.
 */
static inline int json_next_number(struct JSON_Parser *parser,
                                   struct JSON_Token *token)
{
    // current position
    const uint8_t *p = parser->beg;

    int done = 0;
    while (!done && (parser->end - p) > 0)
    {
        // TODO: true "number" state machine
        switch (*p)
        {
            case '0': case '1': case '2': case '3': case '4':
            case '5': case '6': case '7': case '8': case '9':
            case '.': case '-': case '+': case 'e': case 'E':
                ++p;
                break;

            default:
                ++done; // done
        }
    }

    token->beg = parser->beg;
    token->end = p;
    token->type = JSON_NUMBER;
    parser->beg = p;
    return 0; // OK
}


/*
 * json_next() implementation.
 */
int json_next(struct JSON_Parser *parser,
              struct JSON_Token *token)
{
    // check the buffered tokens first
    if (parser->no_tokens)
    {
        struct JSON_Token *b_tok = &parser->tokens[--parser->no_tokens];
        memcpy(token, b_tok, sizeof(*token));
        return 0; // OK, use buffered token
    }

    // ignore whitespaces
    if (!json_skip_ws(parser))
    {
        // no more data to parse
        token->type = JSON_EOF;
        token->beg = parser->beg;
        token->end = parser->end;
        return 0; // OK
    }

    switch (parser->beg[0])
    {
        // begin of JSON object
        case '{':
            token->type = JSON_OBJECT_BEG;
            token->beg = parser->beg++;
            token->end = parser->beg;
            return 0; // OK

        // end of JSON object
        case '}':
            token->type = JSON_OBJECT_END;
            token->beg = parser->beg++;
            token->end = parser->beg;
            return 0; // OK

        // begin of JSON array
        case '[':
            token->type = JSON_ARRAY_BEG;
            token->beg = parser->beg++;
            token->end = parser->beg;
            return 0; // OK

        // end of JSON array
        case ']':
            token->type = JSON_ARRAY_END;
            token->beg = parser->beg++;
            token->end = parser->beg;
            return 0; // OK

        // colon
        case ':':
            token->type = JSON_COLON;
            token->beg = parser->beg++;
            token->end = parser->beg;
            return 0; // OK

        // comma
        case ',':
            token->type = JSON_COMMA;
            token->beg = parser->beg++;
            token->end = parser->beg;
            return 0; // OK

        // parse string ("key" or "value")
        case '\"':
            return json_next_string(parser, token);

        // parse number
        case '-': case '0': case '1' : case '2': case '3' : case '4':
        case '5': case '6': case '7' : case '8': case '9':
            return json_next_number(parser, token);

        // parse "false"
        case 'f':
            if ((parser->end - parser->beg) >= 5 &&
                0 == memcmp(parser->beg+1, "alse", 4)) // "false"
            {
                token->type = JSON_FALSE;
                token->beg = parser->beg;
                token->end = (parser->beg += 5);
                return 0; // OK
            }
            return -1; // failed

        // parse "true"
        case 't':
            if ((parser->end - parser->beg) >= 4 &&
                0 == memcmp(parser->beg, "true", 4)) // "true"
            {
                token->type = JSON_TRUE;
                token->beg = parser->beg;
                token->end = (parser->beg += 4);
                return 0; // OK
            }
            return -1; // failed

        // parse "null"
        case 'n' :
            if ((parser->end - parser->beg) >= 4 &&
                0 == memcmp(parser->beg, "null", 4)) // "null"
            {
                token->type = JSON_NULL;
                token->beg = parser->beg;
                token->end = (parser->beg += 4);
                return 0; // OK
            }
            return -1; // failed
    }

    return -1; // failed, bad token
}


/*
 * json_put_back() implementation.
 */
int json_put_back(struct JSON_Parser *parser,
                  const struct JSON_Token *token)
{
    const int cap = sizeof(parser->tokens) / sizeof(parser->tokens[0]);
    if (parser->no_tokens >= cap)
        return -1; // no space

    struct JSON_Token *b_tok = &parser->tokens[parser->no_tokens++];
    memcpy(b_tok, token, sizeof(*token));
    return 0; // OK, buffered
}


/*
 * json_field_by_name() implementation.
 */
struct JSON_Field* json_field_by_name(struct JSON_Field *field,
                                      const uint8_t *name_beg,
                                      const uint8_t *name_end)
{
    const size_t name_len = (name_end - name_beg);
    if (name_len >= sizeof(field->by_name))
        return 0; // name too big

    // check all sibling fields
    while (field != 0)
    {
        if (0 == memcmp(field->by_name, name_beg, name_len))
            return field;

        field = field->siblings; // goto next
    }

    return 0; // not found
}


/*
 * json_field_by_index() implementation.
 */
struct JSON_Field* json_field_by_index(struct JSON_Field *field,
                                       int index)
{
    // check all sibling fields
    while (field != 0)
    {
        if (field->by_index == index)
            return field;

        field = field->siblings; // goto next
    }

    return 0; // not found
}


/**
 * @brief Create empty field.
 */
static struct JSON_Field* json_field_make(void)
{
    struct JSON_Field *f = (struct JSON_Field*)malloc(sizeof(*f));
    if (f)
        memset(f, 0, sizeof(*f));
    return f;
}


//#define JSON_INDEX_BASE 0
#define JSON_INDEX_BASE 1

/*
 * json_field_parse() implementation.
 */
int json_field_parse(struct JSON_Field **fields,
                     const char *path)
{
    struct JSON_Field *root = 0;
    struct JSON_Field *last = 0;

    while (path && *path)
    {
        // [X] index
        if ('[' == *path)
        {
            char *end = 0;
            long index = strtol(path+1, &end, 0);
            if (!end || end == path+1 || *end != ']')
                return -1; // bad index parsed
            if (index < JSON_INDEX_BASE)
                return -1; // index out of range
            path = end+1; // go to next fields

            // create new field
            struct JSON_Field *f = json_field_make();
            if (!f)
                return -1; // out of memory
            f->by_index = index - JSON_INDEX_BASE;

            if (!root)
                root = f;
            if (last)
                last->children = f;
            last = f;
        }

        // quoted "name"
        else if ('\"' == *path)
        {
            const char *name = ++path; // ignore "
            int done = 0;
            while (!done && *path)
            switch (*path)
            {
                case '\"':
                {
                    size_t len = path - name;
                    ++path;

                    // create new field
                    struct JSON_Field *f = json_field_make();
                    if (!f)
                        return -1; // out of memory
                    if (len >= sizeof(f->by_name))
                        return -1; // name too long
                    memcpy(f->by_name, name, len);
                    f->by_index = -1;

                    if (!root)
                        root = f;
                    if (last)
                        last->children = f;
                    last = f;
                    done = 1; // stop
                } break;

                case '\\':
                    ++path;
                    if (!*path)
                        return -1; // bad escaping
                    ++path;
                    break;

                default:
                    ++path;
                    break;
            }

            if (!done)
                return -1; // bad quoted name
        }

        else if ('.' == *path)
        {
            ++path; // go to next field
        }

        else
        {
            // simple field name
            const char *name = path;
            size_t len = 0;
            int done = 0;
            while (!done && *path)
            switch (*path)
            {
                case '.':
                    ++path; // ignore '.'
                    done = 1; // stop
                    break;

                default:
                    ++path;
                    ++len;
                    break;
            }

            if (len)
            {
                // create new field
                struct JSON_Field *f = json_field_make();
                if (!f)
                    return -1; // out of memory
                if (len >= sizeof(f->by_name))
                    return -1; // name too long
                memcpy(f->by_name, name, len);
                f->by_index = -1;

                if (!root)
                    root = f;
                if (last)
                    last->children = f;
                last = f;
            }
        }
    }

    if (!root)
        return -1; // empty fields
    *fields = root;
    return 0; // OK
}


/*
 * json_field_clone() implementation.
 */
struct JSON_Field* json_field_clone(struct JSON_Field *fields)
{
    if (!fields)
        return 0; // nothing to clone

    struct JSON_Field *res = json_field_make();
    memcpy(res->by_name, fields->by_name, sizeof(res->by_name));
    res->by_index = fields->by_index;
    res->token = fields->token;
    res->data = fields->data;

    res->children = json_field_clone(fields->children);
    res->siblings = json_field_clone(fields->siblings);
    return res;
}


/*
 * json_fields_free() implementation.
 */
void json_field_free(struct JSON_Field *fields)
{
    if (!fields)
        return;

    json_field_free(fields->children);
    json_field_free(fields->siblings);
    free(fields); // release
}


/*
 * json_get() implementation.
 */
int json_get(struct JSON_Parser *parser,
             struct JSON_Field *fields)
{
    // look into JSON object
    struct JSON_Token token;
    if (!!json_next(parser, &token))
        return -1; // failed

    if (JSON_OBJECT_BEG == token.type)
    {
        // iterate over all fields
        for (int i = 0; ; ++i)
        {
            struct JSON_Token key;
            if (!!json_next(parser, &key))
                return -1; // failed
            if (JSON_OBJECT_END == key.type)
                break; // done
            if (JSON_STRING != key.type && JSON_STRING_ESC != key.type)
                return -1; // string key expected

            if (!!json_next(parser, &token))
                return -1; // failed
            if (JSON_COLON != token.type)
                return -1; // colon expected

            if (!!json_next(parser, &token))
                return -1; // failed
            switch (token.type)
            {
                // primitive types
                case JSON_STRING:
                case JSON_STRING_ESC:
                case JSON_NUMBER:
                case JSON_FALSE:
                case JSON_TRUE:
                case JSON_NULL:
                {
                    // TODO: unescape key if JSON_STRING_ESC == key.type
                    struct JSON_Field *sf = json_field_by_name(fields, key.beg, key.end);
                    if (sf) // sub-field matched
                    {
                        // assign current token to sub-field
                        // if (sf.no_field) return -1; ???
                        memcpy(&sf->token, &token, sizeof(token));
                    }
                    else
                    {
                        // token is just ignored...
                    }
                } break;

                // inner OBJECT
                case JSON_OBJECT_BEG:
                {
                    // TODO: unescape key if JSON_STRING_ESC == key.type
                    struct JSON_Field *sf = json_field_by_name(fields, key.beg, key.end);
                    if (sf) // sub-field matched
                    {
                        sf->token.type = JSON_OBJECT;
                        sf->token.beg = token.beg;

                        if (sf->children)
                        {
                            if (!!json_put_back(parser, &token))
                                return -1; // failed, buffer is full
                            if (!!json_get(parser, sf->children))
                                return -1; // failed
                        }
                        else
                        {
                            // ignore the inner OBJECT
                            if (!!json_skip_object(parser))
                                return -1; // failed, bad OBJECT
                        }

                        sf->token.end = parser->beg;
                    }
                    else
                    {
                        // ignore the inner OBJECT
                        if (!!json_skip_object(parser))
                            return -1; // failed, bad OBJECT
                    }
                } break;

                // inner ARRAY
                case JSON_ARRAY_BEG:
                {
                    // TODO: unescape key if JSON_STRING_ESC == key.type
                    struct JSON_Field *sf = json_field_by_name(fields, key.beg, key.end);
                    if (sf) // sub-field matched
                    {
                        sf->token.type = JSON_ARRAY;
                        sf->token.beg = token.beg;

                        if (sf->children)
                        {
                            if (!!json_put_back(parser, &token))
                                return -1; // failed, buffer is full
                            if (!!json_get(parser, sf->children))
                                return -1; // failed
                        }
                        else
                        {
                            // ignore the inner ARRAY
                            if (!!json_skip_array(parser))
                                return -1; // failed, bad ARRAY
                        }

                        sf->token.end = parser->beg;
                    }
                    else
                    {
                        // ignore the inner ARRAY
                        if (!!json_skip_array(parser))
                            return -1; // failed, bad ARRAY
                    }
                } break;

                // bad tokens
                default:
                    return -1; // failed, unexpected token
            }

            // ensure we have "," between fields
            if (!!json_next(parser, &token))
                return -1; // failed
            if (JSON_OBJECT_END == token.type)
                break; // done
            if (JSON_COMMA != token.type)
                return -1; // failed, comma expected
        }
    }
    else if (JSON_ARRAY_BEG == token.type)
    {
        // iterate over all elements
        for (int i = 0; ; ++i)
        {
            if (!!json_next(parser, &token))
                return -1; // failed
            if (JSON_ARRAY_END == token.type)
                break; // done

            switch (token.type)
            {
                // primitive types
                case JSON_STRING:
                case JSON_STRING_ESC:
                case JSON_NUMBER:
                case JSON_FALSE:
                case JSON_TRUE:
                case JSON_NULL:
                {
                    struct JSON_Field *sf = json_field_by_index(fields, i);
                    if (sf) // sub-field matched
                    {
                        // assign current token to sub-field
                        // if (sf.no_field) return -1; ???
                        memcpy(&sf->token, &token, sizeof(token));
                    }
                    else
                    {
                        // token is just ignored...
                    }
                } break;

                // inner OBJECT
                case JSON_OBJECT_BEG:
                {
                    struct JSON_Field *sf = json_field_by_index(fields, i);
                    if (sf) // sub-field matched
                    {
                        sf->token.type = JSON_OBJECT;
                        sf->token.beg = token.beg;

                        if (sf->children)
                        {
                            if (!!json_put_back(parser, &token))
                                return -1; // failed, buffer is full
                            if (!!json_get(parser, sf->children))
                                return -1; // failed
                        }
                        else
                        {
                            // ignore the inner OBJECT
                            if (!!json_skip_object(parser))
                                return -1; // failed, bad OBJECT
                        }

                        sf->token.end = parser->beg;
                    }
                    else
                    {
                        // ignore the inner OBJECT
                        if (!!json_skip_object(parser))
                            return -1; // failed, bad OBJECT
                    }
                } break;

                // inner ARRAY
                case JSON_ARRAY_BEG:
                {
                    struct JSON_Field *sf = json_field_by_index(fields, i);
                    if (sf) // sub-field matched
                    {
                        sf->token.type = JSON_ARRAY;
                        sf->token.beg = token.beg;

                        if (sf->children)
                        {
                            if (!!json_put_back(parser, &token))
                                return -1; // failed, buffer is full
                            if (!!json_get(parser, sf->children))
                                return -1; // failed
                        }
                        else
                        {
                            // ignore the inner ARRAY
                            if (!!json_skip_array(parser))
                                return -1; // failed, bad ARRAY
                        }

                        sf->token.end = parser->beg;
                    }
                    else
                    {
                        // ignore the inner ARRAY
                        if (!!json_skip_array(parser))
                            return -1; // failed, bad ARRAY
                    }
                } break;

                // bad tokens
                default:
                    return -1; // failed, unexpected token
            }

            // ensure we have "," between array elements
            if (!!json_next(parser, &token))
                return -1; // failed
            if (JSON_ARRAY_END == token.type)
                break; // done
            if (JSON_COMMA != token.type)
                return -1; // failed, comma expected
        }
    }
    else
        return -1; // only OBJECT or ARRAY are supported

    return 0; // OK
}
