-- position_workitems

CREATE TABLE position (
    item_id bigserial primary key,
    prev_item bigserial
);

ALTER TABLE ONLY position 
   ADD CONSTRAINT position_item_id_work_items_id_foreign
     FOREIGN KEY (item_id)
     REFERENCES work_items(id)
     ON UPDATE RESTRICT
     ON DELETE RESTRICT;

