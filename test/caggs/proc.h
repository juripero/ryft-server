#ifndef __CAGGS_PROC_H__
#define __CAGGS_PROC_H__

#include <stdint.h>


/**
 * @brief The DATA record reference.
 */
struct RecordRef
{
    uint64_t offset; ///< @brief Record offset in bytes.
    uint64_t length; ///< @brief Record length in bytes.
};


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
int parse_index(const uint8_t *idx_beg,
                const uint8_t *idx_end,
                uint64_t *data_len);


/**
 * @brief Parse record references from INDEX file.
 * @param[in] is_last_chunk Last chunk indicator. If non-zero parses last line even there is no NEWLINE at the end.
 * @param[in] buf Begin of INDEX data.
 * @param[in,out] len Length of INDEX data in bytes on input.
 *                    Length of parsed INDEX data in bytes on output.
 * @param[in] delim_len Delimiter length in bytes.
 * @param[in] data_start Common DATA offset for all record references.
 * @param[in,out] max_data_len Maximum DATA length in bytes on input.
 *                             Actual DATA length in bytes on output.
 * @param[out] records Array with parsed record references.
 * @param[in,out] num_of_records Maximum number of record references on input.
 *                               Number of record references parsed on output.
 * @return Zero on success.
 */
int parse_index_chunk(int is_last_chunk,
                      const uint8_t *buf, uint64_t *len,
                      uint64_t delim_len,
                      uint64_t data_start,
                      uint64_t *max_data_len,
                      struct RecordRef *records,
                      uint64_t *num_of_records);

#endif // __CAGGS_PROC_H__
