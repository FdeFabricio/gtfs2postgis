-- name: create-stops-table
CREATE TABLE stops (
	stop_id VARCHAR(255) NOT NULL PRIMARY KEY,
	stop_code VARCHAR(255),
	stop_name VARCHAR(255) NOT NULL,
	stop_desc VARCHAR(255),
	stop_lat DECIMAL(8,6) NOT NULL,
	stop_lon DECIMAL(9,6) NOT NULL,
	zone_id VARCHAR(255),
	stop_url VARCHAR(511),
	location_type SMALLINT CHECK (location_type BETWEEN 0 AND 1),
	parent_station VARCHAR(255),
	stop_timezone VARCHAR(255),
	wheelchair_boarding SMALLINT CHECK (wheelchair_boarding BETWEEN 0 AND 2)
);
CREATE INDEX stop_lat ON stops (stop_lat);
CREATE INDEX stop_lon ON stops (stop_lon);

-- name: drop-stops-table
DROP TABLE IF EXISTS stops;