
CREATE TABLE airports (
    id BIGINT PRIMARY KEY,           
    ident TEXT,                      
    type TEXT,                       
    name TEXT,
    location GEOMETRY(PointZ, 4326),
    elevation_ft INT,
    iso_country TEXT,
    municipality TEXT,
    icao_code TEXT,
    iata_code TEXT
);

CREATE INDEX idx_airports_geom ON airports USING GIST(location);
CREATE INDEX idx_airports_icao ON airports(icao_code);