package commit

var (
	MergeRequestAdd404OnMissingSeance = MergeRequest{
		Name:   "add-404-on-missing-seance",
		Title:  "Return 404 response when requesting seances for non existent movie",
		Branch: "add-404-on-missing-seance",
		Ref:    "master",
		Commit: Commit{
			Diffs: []string{`
diff --git a/src/api/movies.py b/src/api/movies.py
index ef89db7..98d4dd6 100644
--- a/src/api/movies.py
+++ b/src/api/movies.py
@@ -201,3 +201,8 @@ async def get_movie_seances(request: web.Request):
             # PLACEHOLDER
             #

+            if not is_film_exist['is_exist']:
+                return web.json_response(
+                    {'errors': [{'title': 'Not found', 'detail': f'movie_id {movie_id} not found'}]},
+                    status=404)
+
`},
		},
	}
)
