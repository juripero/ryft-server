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
