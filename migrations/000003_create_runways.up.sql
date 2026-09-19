CREATE TABLE public.runways (
    id                          bigint PRIMARY KEY,
    airport_ref                 bigint,
    airport_ident               text,
    length_ft                   integer,
    width_ft                    integer,
    surface                     text,
    lighted                     boolean,
    closed                      boolean,
    le_ident                    text,
    le_latitude_deg             double precision,
    le_longitude_deg            double precision,
    le_elevation_ft             integer,
    le_heading_deg_t            double precision,
    le_displaced_threshold_ft   integer,
    he_ident                    text,
    he_latitude_deg             double precision,
    he_longitude_deg            double precision,
    he_elevation_ft             integer,
    he_heading_deg_t            double precision,
    he_displaced_threshold_ft   integer,
    geom                        geometry(LineString, 4326),

    CONSTRAINT runways_airport_ref_fkey
        FOREIGN KEY (airport_ref)
        REFERENCES public.airports (id)
        ON DELETE CASCADE
);

CREATE INDEX idx_runways_geom          ON public.runways USING gist (geom);
CREATE INDEX idx_runways_airport_ref   ON public.runways USING btree (airport_ref);
CREATE INDEX idx_runways_airport_ident ON public.runways USING btree (airport_ident);