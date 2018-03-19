#ifndef __CAGGS_STAT_H__
#define __CAGGS_STAT_H__

#include <stdint.h>
#include <stdio.h>


// TODO: abstract classes, etc.
struct Stat {
    uint64_t count; // number of values processed
    double sum, sum2;
    double min, max;
};


/**
 * @brief Initialize existing statistics.
 * @param s Statistics to initialize.
 */
void stat_init(struct Stat *s);


/**
 * @brief Create new statistics.
 *
 * Should be released with stat_free().
 *
 * @return Initialized statistics.
 */
struct Stat* stat_make(void);


/**
 * @brief Create statistics copy.
 * @param base Statistics to clone.
 * @return Cloned statistics. Should be released with stat_free().
 */
struct Stat* stat_clone(const struct Stat *base);


/**
 * @brief Release statistics.
 * @param s Statistics to release.
 */
void stat_free(struct Stat *s);


/**
 * @brief Add floating-point value to the statistics.
 * @param s Statistics to update.
 * @param x Value to add.
 */
void stat_add(struct Stat *s, double x);


/**
 * @brief Merge another statistics.
 * @param to Statistics to update.
 * @param from Statistics to merge from.
 */
void stat_merge(struct Stat *to, const struct Stat *from);


/**
 * @brief Print statistics to the output stream.
 * @param s Statistics to print.
 * @param f Output stream.
 */
void stat_print(const struct Stat *s, FILE *f);


#endif // __CAGGS_STAT_H__
