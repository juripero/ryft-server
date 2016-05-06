#include <stdlib.h>
#include <stdio.h>

#include <pthread.h>

#include <libryftone.h>

// do the fuzzy hamming search and print results
int fhs(const char *tag, const char *query, const char *file, int fuzziness, const char *outfile)
{
	printf("[%s] FHS starting: query:%s, fuzziness:%d\n", tag, query, fuzziness);

	// create Ryft data set
	rol_data_set_t ids = rol_ds_create();

	printf("[%s] adding \"%s\" file to the data set\n", tag, file);

	// add files where to search
	rol_ds_add_file(ids, file);

	printf("[%s] searching (output:\"%s\")...\n", tag, outfile);

	// do search
	rol_data_set_t ods = rol_ds_search_fuzzy_hamming(ids, outfile,
		query, 0, (uint8_t)fuzziness, "\n", NULL, false, NULL);

	// get statistics or error
	if (!rol_ds_has_error_occurred(ods))
	{
		printf("[%s] FHS done: %d matches, %d bytes in %d ms (fabric: %d ms)\n", tag,
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


// search #1
void* fhs_1(void *tag)
{
	fhs((const char*)tag, "(RAW_TEXT CONTAINS \"10\")", "/regression/passengers.txt", 0, "1.dat"); // 12 matches
	pthread_exit(NULL);
	return NULL;
}

// search #2
void* fhs_2(void *tag)
{
	fhs((const char*)tag, "(RAW_TEXT CONTAINS \"20\")", "/regression/passengers.txt", 0, "2.dat"); // 1 match
	pthread_exit(NULL);
	return NULL;
}

// search #3
void* fhs_3(void *tag)
{
	fhs((const char*)tag, "(RAW_TEXT CONTAINS \"555\")", "/regression/passengers.txt", 1, "3.dat"); // 36 matches
	pthread_exit(NULL);
	return NULL;
}

// search #4
void* fhs_4(void *tag)
{
	fhs((const char*)tag, "(RECORD.id CONTAINS \"1003\")", "/regression/*.pcrime", 0, "4.dat"); // 2542 matches
	pthread_exit(NULL);
	return NULL;
}

// search #5
void* fhs_5(void *tag)
{
	fhs((const char*)tag, "(RECORD.id CONTAINS \"1003100\")", "/regression/*.pcrime", 0, "5.dat"); // 9 matches
	pthread_exit(NULL);
	return NULL;
}

// search #6
void* fhs_6(void *tag)
{
	fhs((const char*)tag, "(RECORD.desc CONTAINS \"VEHICLE\")", "/regression/*.pcrime", 0, "6.dat"); // 672 matches
	pthread_exit(NULL);
	return NULL;
}


// entry point
int main(int argc, const char **argv)
{
	if (argc > 1) // serial
	{
		fhs("S1", "(RAW_TEXT CONTAINS \"10\")", "/regression/passengers.txt", 0, "1.dat"); // 12 matches
		fhs("S2", "(RAW_TEXT CONTAINS \"20\")", "/regression/passengers.txt", 0, "2.dat"); // 1 match
		fhs("S3", "(RAW_TEXT CONTAINS \"555\")", "/regression/passengers.txt", 1, "3.dat"); // 36 matches

		fhs("S4", "(RECORD.id CONTAINS \"1003\")", "/regression/*.pcrime", 0, "4.dat"); // 2542 matches
		fhs("S5", "(RECORD.id CONTAINS \"1003100\")", "/regression/*.pcrime", 0, "5.dat"); // 9 matches
		fhs("S6", "(RECORD.desc CONTAINS \"VEHICLE\")", "/regression/*.pcrime", 0, "6.dat"); // 672 matches
	}
	else // concurent
	{
		pthread_t t1, t2, t3, t4, t5, t6;
		pthread_create(&t1, NULL, fhs_1, "X1");
		pthread_create(&t2, NULL, fhs_2, "X2");
		pthread_create(&t3, NULL, fhs_3, "X3");
		pthread_create(&t4, NULL, fhs_4, "X4");
		pthread_create(&t5, NULL, fhs_5, "X5");
		pthread_create(&t6, NULL, fhs_6, "X6");

		pthread_exit(NULL);
	}

	return 0;
}
