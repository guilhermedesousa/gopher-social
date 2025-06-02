ALTER TABLE
    user_invitations
ADD 
    COLUMN expiry timestamp with time zone NOT NULL;