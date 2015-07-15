#include <stdlib.h>
#include <libryftone.h>

// int percentCb(uint8_t percent_complete) {
// 	printf("Percent Complete: %d\n",(int) percent_complete);
// 	return 0;
// }

int main( int argc, char* argv[] ) {
	uint8_t nodesCount = 2;

	rol_data_set_t ds = rol_ds_create_with_nodes(nodesCount);

	bool isFileAdded = true;

	isFileAdded = rol_ds_add_file(ds, "passengers.txt");

	if(!isFileAdded) {
		printf("Is file added passengers.txt -> false");
		return 1;
	}


	rol_data_set_t searchResultsDs = rol_ds_search_exact(
		ds,
     	"passengers-results.txt", // results file
     	"(RAW_TEXT CONTAINS \"lle\")", //query
     	0, //surrounding width
     	"\n", // delimeter string
     	NULL, // index results file
     	NULL // precentage callback
    );

    if (rol_ds_has_error_occurred(searchResultsDs)) {
		printf("Error occurred during search operation.\n");
		printf("Error string:%s\n", rol_ds_get_error_string(searchResultsDs));
	}

	rol_ds_delete(&ds);
	rol_ds_delete(&searchResultsDs);

	return 0;

}
