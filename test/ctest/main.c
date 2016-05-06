#include <stdlib.h>
#include <stdio.h>

#include <libryftone.h>

// do the fuzzy hamming search and print results
int fhs(const char *tag, const char *query, const char *file, int fuzziness, const char *outfile)
{
	printf("[%s]  FHS: query:%s, file:\"%s\", fuzziness:%d, output:\"%s\"\n",
		tag, query, file, fuzziness, outfile);

	// create Ryft data set
	rol_data_set_t ids = rol_ds_create();

	// add files where to search
	rol_ds_add_file(ids, file);

	// do search
	rol_data_set_t ods = rol_ds_search_fuzzy_hamming(ids, outfile,
		query, 0, (uint8_t)fuzziness, "\n", NULL, false, NULL);

	// get statistics or error
	if (!rol_ds_has_error_occurred(ods))
	{
		printf("[%s] DONE: %d matches, %d bytes in %d ms (fabric: %d ms)\n", tag,
			(int)rol_ds_get_total_matches(ods),
			(int)rol_ds_get_total_bytes_processed(ods),
			(int)rol_ds_get_execution_duration(ods),
			(int)rol_ds_get_fabric_execution_duration(ods));
	}
	else
	{
		printf("[%s] ERROR: %s\n", tag, rol_ds_get_error_string(ods));
	}

	// cleanup
	rol_ds_delete(&ods);
	rol_ds_delete(&ids);

	return 0; // OK
}

// entry point
int main()
{
	if (1) // sequence
	{
		fhs("S1", "(RAW_TEXT CONTAINS \"10\")", "/regression/passengers.txt", 0, "1.dat"); // 12 matches
		fhs("S2", "(RAW_TEXT CONTAINS \"20\")", "/regression/passengers.txt", 0, "2.dat"); // 1 match
		fhs("S3", "(RAW_TEXT CONTAINS \"555\")", "/regression/passengers.txt", 1, "3.dat"); // 36 matches

		fhs("S4", "(RECORD.id CONTAINS \"1003\")", "/regression/*.pcrime", 0, "4.dat"); // 2542 matches
		fhs("S5", "(RECORD.id CONTAINS \"1003100\")", "/regression/*.pcrime", 0, "5.dat"); // 9 matches
		fhs("S6", "(RECORD.desc CONTAINS \"VEHICLE\")", "/regression/*.pcrime", 0, "6.dat"); // 672 matches
	}

	return 0;
}
