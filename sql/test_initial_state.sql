TRUNCATE TABLE "users" CASCADE;
TRUNCATE TABLE "roles" CASCADE;
TRUNCATE TABLE "classes" CASCADE;
TRUNCATE TABLE "age_ranges" CASCADE;
TRUNCATE TABLE "responsible_of" CASCADE;
TRUNCATE TABLE "children" CASCADE;
TRUNCATE TABLE "age_ranges" CASCADE;

INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id1','arthur@gmail.com','Arthur','Gustin','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id2','vinu@gmail.com','Vinu','Singh','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id3','john@gmail.com','John','John','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id4','estree@gmail.com','Estree','Delacour','F','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com');
INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id5','anna@gmail.com','Anna','Melnychuk','F','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com');
INSERT INTO "roles" ("user_id","role") VALUES ('id1', 'admin');
INSERT INTO "roles" ("user_id","role") VALUES ('id2', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id3', 'officemanager');
INSERT INTO "roles" ("user_id","role") VALUES ('id4', 'teacher');
INSERT INTO "roles" ("user_id","role") VALUES ('id5', 'adult');

INSERT INTO "age_ranges" ("age_range_id","stage","min","min_unit","max","max_unit") VALUES ('agerangeid-1','infant','3','M','12','M');
INSERT INTO "age_ranges" ("age_range_id","stage","min","min_unit","max","max_unit") VALUES ('agerangeid-2','toddlers','12','M','18','M');
INSERT INTO "classes" ("class_id","age_range_id","name","description","image_uri") VALUES ('classid-1', 'agerangeid-1', 'infant class','infant description','gs://foo/bar.jpg');
INSERT INTO "classes" ("class_id","age_range_id","name","description","image_uri") VALUES ('classid-2', 'agerangeid-2', 'toddlers class','toddlers description','gs://foo/bar.jpg');

INSERT INTO "children" ("child_id","first_name","last_name","birth_date","start_date","gender","image_uri","notes") VALUES ('childid-1','Arthur','Gustin','1992-10-13T00:00:00Z','2018-03-28T00:00:00Z','M','gs://foo/bar.jpg', 'some special notes');
INSERT INTO "allergies" ("allergy_id","child_id","allergy") VALUES ('allergyid-1','childid-1','tomato');
INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('id5','childid-1','father');
INSERT INTO "special_instructions" ("child_id","instruction") VALUES ('childid-1','this boy always sleeps please keep him awaken');