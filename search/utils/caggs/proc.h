/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

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
    struct JSON_Field *field_tree; ///< @brief Field tree to search for.

    struct JSON_Field **fields; ///< @brief Target fields.
    struct Stat **stats;        ///< @brief Final statistics (per target field).
    int n_fields;               ///< @brief Number of target fields.

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
