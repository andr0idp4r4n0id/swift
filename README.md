Swift fuzzes the query string parameters within the urls given through stdin, with values that figure out the template being used.

Usage; For non blind ssti; cat list_of_urls | sheep -concurrency 50 | mirror -concurrency 50 | swift -concurrency 50
