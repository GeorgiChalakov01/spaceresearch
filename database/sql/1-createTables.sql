BEGIN;

create table users (
	id serial primary key,
	email text not null,
	passwordHash text not null,
	isAdmin boolean not null default false
);

COMMIT;
