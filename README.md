# gtfs2postgis

Properties of fields:

It seems that there is no convention regarding the length of fields. Whilst GTFS website doesn't mention it, transportation organisations diverge:
[transitwiki](https://www.transitwiki.org/TransitWiki/images/e/e7/GTFS+_Additional_Files_Format_Ver_1.7.pdf) mentions 15 characters for `stop_id` while [MBTA](https://mbta.com/developers/gtfs-documentation#stops) states 45.
For this reason I adopted a very high limit of 255 characters.