CREATE OR REPLACE FUNCTION adds_order() RETURNS void as $$
-- adds_order() function limits order to space
	DECLARE
	i integer=1000;
	r RECORD;
	a RECORD;
	xyz CURSOR FOR SELECT id, execution_order from work_items;
	abc CURSOR FOR SELECT id from spaces;

	BEGIN
		open abc;
		FOR a in FETCH ALL FROM abc
			LOOP
				open xyz;
				FOR r in FETCH ALL FROM xyz LOOP
						UPDATE work_items set execution_order=i where id=r.id AND space_id=a.id;
						i := i+1000;
				END LOOP;
				close xyz;
			END LOOP;
			close abc;
	END $$

DO $$ BEGIN
	PERFORM adds_order();
	DROP FUNCTION adds_order();
END $$

