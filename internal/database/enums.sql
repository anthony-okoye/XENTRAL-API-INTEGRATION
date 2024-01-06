CREATE TYPE user_role AS ENUM (
    'admin',
    'customer'
);

CREATE TYPE user_salutation AS ENUM (
    'Herr',
    'Frau'
);

CREATE TYPE user_type AS ENUM (
    'Privat',
    'Gewerblich'
);


CREATE TYPE payment_method AS ENUM (
    'card',
    'bank'
);

CREATE TYPE order_status AS ENUM (
    'in_progress',
    'failed',
    'finished'
);

CREATE TYPE delivery_status AS ENUM (
    'open',
    'sent',
    'cancelled'
);

CREATE TYPE payment_status AS ENUM (
    'pending',
    'paid',
    'failed'
);