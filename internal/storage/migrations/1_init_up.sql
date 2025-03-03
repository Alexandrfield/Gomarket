CREATE TABLE if NOT EXISTS Users (id SERIAL PRIMARY KEY, login text, passwd text, allPoints double precision, usedPoints double precision);
CREATE TABLE if NOT EXISTS Orders (id SERIAL PRIMARY KEY, numer bigint, polsak int, status text, points double precision, upload timestamp);
CREATE TABLE if NOT EXISTS Used (id SERIAL PRIMARY KEY, numer bigint, polsak int, sum double precision, upload timestamp);
