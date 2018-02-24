#include "proc.h"
#include "misc.h"

#include <string.h>
#include <stdlib.h>

/*
 * parse_index() implementation.
 */
int parse_index(const uint8_t *idx_beg,
                const uint8_t *idx_end,
                uint64_t *data_len)
{
    const int COMMA = ',';

    //extern void print_buf(const void*, const void*);
    //printf("parsing INDEX: "); print_buf(idx_beg, idx_end-1); printf("\n");

    // find ",fuzziness"
    const uint8_t *c4 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                idx_end - idx_beg);
    if (!c4)
    {
        verr3("parse_index: no \",fuzziness\" found\n");
        return -1; // no ",fuzziness" found
    }

    // find ",length"
    const uint8_t *c3 = (const uint8_t*)memrchr(idx_beg, COMMA,
                                                c4 - idx_beg);
    if (!c3)
    {
        verr3("parse_index: no \",length\" found\n");
        return -2; // no ",length" found
    }

    uint8_t *end = 0;
    *data_len = strtoull((const char*)c3+1, // +1 to skip comma
                         (char**)&end, 10);
    if (end != c4)
    {
        verr3("parse_index: failed to parse length\n");
        return -3; // failed to parse length
    }

    return 0; // OK
}



/*
 * parse_index_chunk() implementation.
 */
int parse_index_chunk(int is_last_chunk,
                      const uint8_t *buf, uint64_t *len_,
                      uint64_t delim_len,
                      uint64_t data_start,
                      uint64_t *max_data_len_,
                      struct RecordRef *records,
                      uint64_t *num_of_records_)
{
    int64_t len = *len_; // INDEX buffer remain in bytes

    // total DATA length in bytes
    const uint64_t max_data_len = *max_data_len_;
    uint64_t data_len = 0;

    // number of indices/record references parsed
    const uint64_t max_count = *num_of_records_;
    uint64_t count = 0;

    int res = 0;

    // read INDEX line by line
    while (0 < len && count < max_count)
    {
        // try to find the NEWLINE '\n' character
        const uint8_t *eol = (const uint8_t*)memchr(buf, '\n', len);
        if (!eol && !is_last_chunk)
        {
            res = 1;
            break; // done, leave the last INDEX (part of it) to the next chunk
        }

        const uint8_t *next = eol ? (eol+1) : (buf+len);

        uint64_t d_len = 0;
        if (!!parse_index(buf, next, &d_len))
        {
            verr1("ERROR: failed to parse INDEX\n"); // TODO: add "at" information from idx_beg
            return -2; // failed
        }

        if (data_len + d_len + delim_len > max_data_len)
            break; // no space to save data

        records[count].offset = data_start + data_len;
        records[count].length = d_len;
        count += 1;

        // go to next record...
        data_len += d_len + delim_len;
        len -= (next - buf);
        buf = next;
    }

    // results
    *max_data_len_ = data_len;
    *num_of_records_ = count;
    *len_ -= len;

    return res; // OK
}
