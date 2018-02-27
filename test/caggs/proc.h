#ifndef __CAGGS_PROC_H__
#define __CAGGS_PROC_H__

#include "conf.h"

#include <stdint.h>
#include <stdio.h>


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


/**
 * @brief Work related parameters and results.
 */
struct Work
{
    struct JSON_Field *field; ///< @brief Field to search for.

    struct Stat *stat; ///< @brief Final statistics.

    struct XProc *xproc; ///< @brief XProc processing units.
    int n_xproc; ///< @brief Number of XProc units.

    int xproc_started; /// <@brief XProc "active" flag.
    int64_t xproc_start; ///< @brief The last iteration start time, microseconds.
};


/**
 * @brief Create new work structure.
 * @param cfg Configuration.
 * @return NULL on failure.
 */
struct Work* work_make(const struct Conf *cfg);


/**
 * @brief Release work structure.
 * @param w Work to release.
 */
void work_free(struct Work *w);


/**
 * @brief Do start work processing.
 *
 * DATA buffer and record references should not be
 * used or changed until work is done.
 *
 * @param w Work to process.
 * @param data_buf Begin of DATA buffer.
 * @param records Record references.
 * @param num_of_records Number of record references.
 * @return Zero on success.
 */
int work_do_start(struct Work *w, const uint8_t *data_buf,
                  const struct RecordRef *records,
                  uint64_t num_of_records);


/**
 * @brief Do join the work processing.
 * @param w Work to join.
 * @return Zero on success.
 */
int work_do_join(struct Work *w);


/**
 * @brief Print work results.
 * @param w Work to get results from.
 * @param f Output stream.
 */
void work_print(struct Work *w, FILE *f);

#endif // __CAGGS_PROC_H__
