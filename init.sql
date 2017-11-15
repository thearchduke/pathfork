drop table if exists tbl_subsection;
drop table if exists tbl_section CASCADE;
drop table if exists tbl_snippet;
drop table if exists tbl_work CASCADE;
drop table if exists tbl_character;
drop table if exists tbl_setting;
drop table if exists tbl_thing;
drop table if exists tbl_user CASCADE;
drop table if exists r_works_characters;
drop table if exists r_sections_characters;
drop table if exists r_works_settings;
drop table if exists r_sections_settings;
drop table if exists r_works_things;
drop table if exists r_sections_things;
drop table if exists r_settings_things;
drop table if exists r_settings_characters;
drop table if exists r_characters_things;

create table tbl_user(
email varchar(256) primary key,
pw varchar(64) not null,
verified BOOLEAN DEFAULT FALSE
);

create table tbl_work(
work_id serial primary key,
title text not null,
blurb text,
user_email text not null,
word_count integer not null default 0,
foreign key (user_email) references tbl_user(email)
	ON DELETE CASCADE
);

create table tbl_section(
section_id serial primary key,
title text not null,
blurb text,
body text,
order integer,
is_snippet boolean default false,
work_id integer not null,
user_email text not null,
word_count integer not null default 0,
foreign key (work_id) references tbl_work(work_id)
	ON DELETE CASCADE,
foreign key (user_email) references tbl_user(email)
	ON DELETE CASCADE
);

/*
create table tbl_subsection(
subsection_id serial primary key,
title text not null,
blurb text,
body text,
section_id integer not null,
user_email text not null,
foreign key (section_id) references tbl_section(section_id),
foreign key (user_email) references tbl_user(email)
);
*/

/*
create table tbl_snippet(
snippet_id serial primary key,
title text not null,
blurb text,
body text,
work_id integer,
user_email text not null,
foreign key (work_id) references tbl_work(work_id)
	ON DELETE CASCADE,
foreign key (user_email) references tbl_user(email)
	ON DELETE CASCADE
);
*/

create table tbl_character(
character_id serial primary key,
name text not null,
blurb text,
body text,
user_email text not null,
foreign key (user_email) references tbl_user(email)
	ON DELETE CASCADE
);

create table r_sections_characters(
section_id serial not null,
character_id integer not null,
PRIMARY KEY (section_id, character_id),
FOREIGN KEY (section_id) references tbl_section(section_id)
	ON DELETE CASCADE,
FOREIGN KEY (character_id) references tbl_character(character_id)
	ON DELETE CASCADE
);

create table r_works_characters(
work_id serial not null,
character_id integer not null,
PRIMARY KEY (work_id, character_id),
FOREIGN KEY (work_id) references tbl_work(work_id)
	ON DELETE CASCADE,
FOREIGN KEY (character_id) references tbl_character(character_id)
	ON DELETE CASCADE
);

create table tbl_setting(
setting_id serial primary key,
name text not null,
blurb text,
body text,
user_email text not null,
foreign key (user_email) references tbl_user(email)
	ON DELETE CASCADE
);

create table r_sections_settings(
section_id serial not null,
setting_id serial not null,
PRIMARY KEY (section_id, setting_id),
FOREIGN KEY (section_id) references tbl_section(section_id)
	ON DELETE CASCADE,
FOREIGN KEY (setting_id) references tbl_setting(setting_id)
	ON DELETE CASCADE
);

create table r_works_settings(
work_id serial not null,
setting_id serial not null,
PRIMARY KEY (work_id, setting_id),
FOREIGN KEY (work_id) references tbl_work(work_id)
	ON DELETE CASCADE,
FOREIGN KEY (setting_id) references tbl_setting(setting_id)
	ON DELETE CASCADE
);

/*
create table tbl_thing(
thing_id serial primary key,
name text not null,
blurb text,
user_email text not null,
foreign key (user_email) references tbl_user(email)
);

create table r_sections_things(
section_id integer not null,
thing_id integer not null,
PRIMARY KEY (section_id, thing_id)
);

create table r_works_things(
work_id integer not null,
thing_id integer not null,
PRIMARY KEY (work_id, thing_id)
);

create table r_characters_things(
character_id integer not null,
thing_id integer not null,
PRIMARY KEY (character_id, thing_id)
);

create table r_settings_things(
setting_id integer not null,
thing_id integer not null,
PRIMARY KEY (setting_id, thing_id)
);
*/

create table r_settings_characters(
setting_id integer not null,
character_id integer not null,
PRIMARY KEY (setting_id, character_id),
FOREIGN KEY (setting_id) references tbl_setting(setting_id)
	ON DELETE CASCADE,
FOREIGN KEY (character_id) references tbl_character(character_id)
	ON DELETE CASCADE
);

create unique index ix_characters_works on r_works_characters (character_id, work_id);
create unique index ix_settings_works on r_works_settings (setting_id, work_id);
create unique index ix_characters_sections on r_sections_characters (character_id, section_id);
create unique index ix_settings_sections on r_sections_settings (setting_id, section_id);
/*
create unique index ix_things_sections on r_sections_things (thing_id, section_id);
create unique index ix_things_works on r_works_things (thing_id, work_id);
*/

create index ix_work_email on tbl_work (user_email);
create index ix_character_email on tbl_character (user_email);
create index ix_setting_email on tbl_setting (user_email);
/*create index ix_thing_email on tbl_thing (user_email);*/


/*
INSERT INTO tbl_user(email, pw) values ('tynanburke@gmail.com', 'password');
INSERT INTO tbl_work(title, user_email) values ('my first novel!', 'tynanburke@gmail.com');
INSERT INTO tbl_section(title, user_email, work_id) values ('a title', 'tynanburke@gmail.com', 2);
insert into tbl_section(title, work_id, user_email) values ('section 2', 1, 'tynanburke@gmail.com');
insert into tbl_section(title, blurb, body, work_id, user_email) values ('section 3', 'In which there is discontent', 'Now is the winter of our discontent', 1, 'tynanburke@gmail.com');
insert into tbl_character(name, user_email) values ('cornelius', 'tynanburke@gmail.com');
insert into r_works_characters values (1, 1);
*/
