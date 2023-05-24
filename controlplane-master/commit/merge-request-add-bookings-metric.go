package commit

//import "github.com/xanzy/go-gitlab"

const (
	BookingsMetricIncrementDiff = `
diff --git a/src/api/bookings.py b/src/api/bookings.py
index f8e45a8..6177773 100644
--- a/src/api/bookings.py
+++ b/src/api/bookings.py
@@ -5,6 +5,7 @@ from aiohttp import web
 from aiomysql import DictCursor
 from pymysql.err import IntegrityError
 from src.misc import add_optionally_slashed_route, validatable
+from os import getenv


 class Bookings(web.Application):
@@ -92,6 +93,8 @@ async def create_booking(app, payload, additional_headers):
                             status=404)
                     return web.json_response({'errors': [{'title': err.args[1]}]}, status=400)

+        city = getenv('PROVIDER_CITY', 'unknown')
+        app['success_bookings_total'].add({'app': 'provider_backend', 'city': city}, len(bookings_id))
         #
         # PLACEHOLDER
         #
`
)

var (
	MergeRequestAddBookingsMetric = MergeRequest{
		Name:   "add-bookings-metric",
		Title:  "Add business metric to count successful bookings",
		Branch: "add-bookings-metric",
		Ref:    "master",
		Commit: Commit{
			Diffs: []string{BookingsMetricIncrementDiff},
		},
	}
)
