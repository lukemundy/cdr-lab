-- The migrations create these roles, thus the lab runs this file after them, at each start.

GRANT register_app TO register;
-- Roles are global to the server, so the migrations of wattle-bank, which runs with no profile, create the role also
-- for the banks of the all-banks profile.
GRANT dataholder_app TO wattle_bank, redgum_bank, harbour_bank, kestrel_bank;
GRANT datarecipient_app TO sorrel;
