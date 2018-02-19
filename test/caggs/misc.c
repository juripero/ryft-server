#include "misc.h"

#include <stdlib.h>
#include <string.h>

/*
 * be quiet by default.
 */
int verbose = 0;


/*
 * parse_len() implementation.
 */
int parse_len(const char *str, int *len)
{
    if (!str)
        return -1; // no string provided

    // parse bytes
    char *end = 0;
    double b = strtod(str, &end);

    double x; // scale
    if (0 == strcmp(end, "G") || 0 == strcmp(end, "g"))
        x = 1024*1024*1024;
    else if (0 == strcmp(end, "M") || 0 == strcmp(end, "m"))
        x = 1024*1024;
    else if (0 == strcmp(end, "K") || 0 == strcmp(end, "k"))
        x = 1024;
    else if (0 == strcmp(end, "B") || 0 == strcmp(end, "b") || 0 == strcmp(end, ""))
        x = 1;
    else
        return -2; // suffix is unknown!

    // save length (bytes)
    if (len) *len = (int)(b*x + 0.5); // round

    return 0; // OK
}
