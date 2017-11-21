#ifndef __CAGGS_JSON_H__
#define __CAGGS_JSON_H__

#include <stddef.h>
#include <stdint.h>


/**
 * @brief JSON token type.
 */
enum JSON_TokenType {
    JSON_EOF,

    JSON_COLON,         // :
    JSON_COMMA,         // ,

    JSON_OBJECT,        // { ... }
    JSON_OBJECT_BEG,    // {
    JSON_OBJECT_END,    // }

    JSON_ARRAY,         // [ ... ]
    JSON_ARRAY_BEG,     // [
    JSON_ARRAY_END,     // ]

    JSON_STRING,        // "val"
    JSON_STRING_ESC,    // "val-\u1234" (with escaped symbols)
    JSON_NUMBER,        // 123.456
    JSON_FALSE,         // false
    JSON_TRUE,          // true
    JSON_NULL           // null
};


/**
 * @brief JSON token.
 *
 * Contains type and pointers to corresponding data.
 */
struct JSON_Token
{
    const uint8_t *beg; ///< @brief Begin of element.
    const uint8_t *end; ///< @brief End of element.
    enum JSON_TokenType type; ///< @brief Token type.
};


/**
 * @brief JSON parser.
 *
 * Contains parser state.
 */
struct JSON_Parser {
    const uint8_t *beg; ///< @brief Begin of JSON data.
    const uint8_t *end; ///< @brief End of JSON data.
    int state; ///< @brief Parser's state.

    // buffered tokens
    struct JSON_Token tokens[32];   ///< @brief Buffer of tokens.
    int            no_tokens;       ///< @brief Number of buffered tokens.
};


/**
 * @brief Initialize JSON parser.
 *
 * @param[in] parser JSON parser to initialize.
 * @param[in] json_beg Begin of JSON data.
 * @param[in] json_end End of JSON data.
 */
void json_init(struct JSON_Parser *parser,
               const void *json_beg,
               const void *json_end);


/**
 * @brief Get next token.
 *
 * @param[in] parser JSON parser.
 * @param[out] token The parsed token.
 * @return Zero on success.
 */
int json_next(struct JSON_Parser *parser,
              struct JSON_Token *token);


/**
 * @brief Put the token back to parser.
 * @param parser JSON parser.
 * @param token The token to revert.
 * @return Zero on success.
 */
int json_put_back(struct JSON_Parser *parser,
                  const struct JSON_Token *token);


/**
 * @brief JSON fields to access data.
 */
struct JSON_Field
{
    int by_index;       ///< @brief Array index or -1 (by field name).
    char by_name[64];   ///< @brief Field name. Up to 63 bytes.

    /// @brief Corresponding JSON token.
    struct JSON_Token token;

    // sub-fields
    int no_fields;                  ///< @brief Number of sub-fields.
    struct JSON_Field* fields[32];  ///< @brief Array of sub-fields.
};

/**
 * @brief Special index to indicate access "by name".
 */
static const int JSON_FIELD_BY_NAME = -1;


/**
 * @brief Parse the JSON field.
 * @param f Parsed field.
 * @param path Path to parse.
 * @return Zero on success.
 */
int json_field_parse(struct JSON_Field **f,
                     const char *path);


/**
 * @brief Release the JSON field.
 * @param f Field to release.
 */
void json_field_free(struct JSON_Field *f);


/**
 * @brief Get JSON data.
 *
 * Validate JSON and gets corresponding fields.
 *
 * @param[in] parser JSON parser.
 * @param[int] field Fields to get.
 * @return Zero on success.
 */
int json_get(struct JSON_Parser *parser,
             struct JSON_Field *field);

#endif // __CAGGS_JSON_H__
