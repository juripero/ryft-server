#include "json.h"

#include <string.h>


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
static inline int json_skip_ws(struct JSON_Parser *parser)
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


    return 0; // OK
}
