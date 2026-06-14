-- no transaction needed tools wraps automatically

create table flights (
    flight_id serial primary key,
    callsign varchar(20),
    is_anomaly boolean default false,
    raw_file_path text
);
	
create index idx_flights_callsign on flights(callsign);
