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
