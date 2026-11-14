-- Lab-only users and passwords. The PostgreSQL image runs this file one time, when the volume is empty.
-- The owner users have CREATEROLE, because the migrations of each service create its roles.
-- All four preset banks have a database, also when the all-banks profile is off, so that a lab can turn it on later.

CREATE ROLE register_owner LOGIN CREATEROLE PASSWORD 'register_owner';
CREATE ROLE register LOGIN PASSWORD 'register';
CREATE DATABASE register OWNER register_owner;
REVOKE ALL ON DATABASE register FROM PUBLIC;
GRANT CONNECT ON DATABASE register TO register;

CREATE ROLE wattle_bank_owner LOGIN CREATEROLE PASSWORD 'wattle_bank_owner';
CREATE ROLE wattle_bank LOGIN PASSWORD 'wattle_bank';
CREATE DATABASE wattle_bank OWNER wattle_bank_owner;
REVOKE ALL ON DATABASE wattle_bank FROM PUBLIC;
GRANT CONNECT ON DATABASE wattle_bank TO wattle_bank;

CREATE ROLE redgum_bank_owner LOGIN CREATEROLE PASSWORD 'redgum_bank_owner';
CREATE ROLE redgum_bank LOGIN PASSWORD 'redgum_bank';
CREATE DATABASE redgum_bank OWNER redgum_bank_owner;
REVOKE ALL ON DATABASE redgum_bank FROM PUBLIC;
GRANT CONNECT ON DATABASE redgum_bank TO redgum_bank;

CREATE ROLE harbour_bank_owner LOGIN CREATEROLE PASSWORD 'harbour_bank_owner';
CREATE ROLE harbour_bank LOGIN PASSWORD 'harbour_bank';
CREATE DATABASE harbour_bank OWNER harbour_bank_owner;
REVOKE ALL ON DATABASE harbour_bank FROM PUBLIC;
GRANT CONNECT ON DATABASE harbour_bank TO harbour_bank;

CREATE ROLE kestrel_bank_owner LOGIN CREATEROLE PASSWORD 'kestrel_bank_owner';
CREATE ROLE kestrel_bank LOGIN PASSWORD 'kestrel_bank';
CREATE DATABASE kestrel_bank OWNER kestrel_bank_owner;
REVOKE ALL ON DATABASE kestrel_bank FROM PUBLIC;
GRANT CONNECT ON DATABASE kestrel_bank TO kestrel_bank;
