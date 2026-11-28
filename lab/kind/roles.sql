-- The pod of each service runs its migrations and its seed with no grant step between them, so the roles exist first.

CREATE ROLE register_app NOLOGIN;
CREATE ROLE dataholder_app NOLOGIN;
CREATE ROLE datarecipient_app NOLOGIN;
