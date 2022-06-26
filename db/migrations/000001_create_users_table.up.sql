create table if not exists users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(50) unique not null,
  password_hash VARCHAR(50) not null,
  email VARCHAR(120) unique not null
)
