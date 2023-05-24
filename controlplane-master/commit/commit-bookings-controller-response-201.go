package commit

var (
	CommitCreateBookingsResponse201 = Commit{
		Name:    "bookings-controller-response-201",
		Message: "Change the successful booking HTTP response status to 201 (Created)",
		Branch:  "master",
		Diffs: []string{`
diff --git a/src/api/bookings.py b/src/api/bookings.py
index f30d3b3..3d9f91b 100644
--- a/src/api/bookings.py
+++ b/src/api/bookings.py
@@ -97,4 +97,5 @@ async def create_booking(app, payload, additional_headers):
         #

         return web.json_response({'data': bookings_id},
-                                 headers={**app['default_headers'], **additional_headers})
+                                 headers={**app['default_headers'], **additional_headers},
+                                 status=201)
`,
		},
	}
)
