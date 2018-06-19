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

/** @brief
 * @brief Application configuration.
 */
#ifndef __CAGGS_CONF_H__
#define __CAGGS_CONF_H__

#include <stdint.h>


/**
 * @brief Application configuration.
 */
struct Conf
{
    const char *idx_path;   ///< @brief Path to INDEX file.
    const char *dat_path;   ///< @brief Path to DATA file.

    const char **fields;    ///< @brief Fields to access data.
    int n_fields;           ///< @brief Number of fields.

    uint64_t header_len;    ///< @brief DATA header in bytes.
    uint64_t delim_len;     ///< @brief DATA delimiter in bytes.
    uint64_t footer_len;    ///< @brief DATA footer in bytes.

    uint64_t idx_chunk_size; ///< @brief INDEX processing chunk size in bytes.
    uint64_t dat_chunk_size; ///< @brief DATA processing chunk size in bytes.
    uint64_t rec_per_chunk; ///< @brief Maximum number of records per DATA chunk.

    int concurrency;        ///< @brief Number of processing threads.
};


/**
 * @brief Parse configuration from command line.
 * @param[out] cfg Configuration parsed.
 * @param[in] argc Number of command line arguments.
 * @param[in] argv List of command line arguments.
 * @return Zero on success.
 */
int conf_parse(struct Conf *cfg, int argc,
               const char *argv[]);


/**
 * @brief Release configuration resources.
 * @param[in] cfg Configuration to release.
 * @return Zero on success.
 */
int conf_free(struct Conf *cfg);


/**
 * @brief Print configuration to STDOUT.
 * @param[in] cfg Configuration to print.
 */
void conf_print(const struct Conf *cfg);


#endif // __CAGGS_CONF_H__
