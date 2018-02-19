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
    char by_name[64];   ///< @brief Field name. Up to 63 bytes.
    int by_index;       ///< @brief Array index or -1 (by field name).

    /// @brief Corresponding JSON token.
    struct JSON_Token token;

    /// @brief Custom data.
    void *data;

    struct JSON_Field *children;  ///< @brief List of child fields.
    struct JSON_Field *siblings;  ///< @brief List of sibling fields.
};


/**
 * @brief Parse the fields tree from path.
 * @param[in,out] fields Head of parsed fields tree.
 * @param[in] path Path to parse, for example "foo.bar".
 * @return Zero on success.
 */
int json_field_parse(struct JSON_Field **field,
                     const char *path);


/**
 * @brief Clone the fields tree.
 * @param[in] fields Head of fields tree to clone.
 * @return Cloned fields tree.
 */
struct JSON_Field* json_field_clone(struct JSON_Field *fields);


/**
 * @brief Release the JSON fields tree.
 * @param[in] fields Head of fields tree.
 */
void json_field_free(struct JSON_Field *fields);


/**
 * @brief Find sibling field by name.
 * @param[in] field Head of fields tree.
 * @param[in] name_beg Begin of name to search.
 * @param[in] name_end End of name to search.
 * @return Sibling field or `NULL` if not found.
 */
struct JSON_Field* json_field_by_name(struct JSON_Field *field,
                                      const uint8_t *name_beg,
                                      const uint8_t *name_end);


/**
 * @brief Find sibling field by index.
 * @param[in] field Head of fields tree.
 * @param[in] index Index to search.
 * @return Sibling field or `NULL` if not found.
 */
struct JSON_Field* json_field_by_index(struct JSON_Field *field,
                                       int index);


/**
 * @brief Get JSON data.
 *
 * Validate JSON and gets corresponding fields.
 *
 * @param[in] parser JSON parser.
 * @param[in] fields Fields tree to get.
 * @return Zero on success.
 */
int json_get(struct JSON_Parser *parser,
             struct JSON_Field *fields);

#endif // __CAGGS_JSON_H__
