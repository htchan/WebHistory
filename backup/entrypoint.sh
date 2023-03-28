mkdir -p /backup/$(date '+%Y-%m-%d')
PGPASSWORD=$PSQL_PASSWORD pg_dump -h $PSQL_HOST -p $PSQL_PORT -U $PSQL_USER -d $PSQL_NAME --schema-only | sed '/^SET\|^SELECT\|^-- Dumped/d' | sed '/^$/{N;/^\n$/d;}' > /backup/$(date '+%Y-%m-%d')/schema.sql;
PGPASSWORD=$PSQL_PASSWORD pg_dump -h $PSQL_HOST -p $PSQL_PORT -U $PSQL_USER -d $PSQL_NAME --data-only | sed '/^SET\|^SELECT\|^-- Dumped/d' | sed '/^$/{N;/^\n$/d;}' > /backup/$(date '+%Y-%m-%d')/data.sql;

### backup by table ###
# tables=$(PGPASSWORD=$PSQL_PASSWORD psql -h $PSQL_HOST -p $PSQL_PORT -U $PSQL_USER -d $PSQL_NAME -t -c "Select table_name From information_schema.tables Where table_type='BASE TABLE' and table_schema='public'")

# for table in $tables;
#   do 
#   PGPASSWORD=$PSQL_PASSWORD pg_dump -h $PSQL_HOST -p $PSQL_PORT -U $PSQL_USER -d $PSQL_NAME -t $table --schema-only | sed '/^SET\|^SELECT\|^-- Dumped/d' | sed '/^$/{N;/^\n$/d;}' > /backup/$(date '+%Y-%m-%d')/$table.schema.sql;
#   PGPASSWORD=$PSQL_PASSWORD pg_dump -h $PSQL_HOST -p $PSQL_PORT -U $PSQL_USER -d $PSQL_NAME -t $table --data-only | sed '/^SET\|^SELECT\|^-- Dumped/d' | sed '/^$/{N;/^\n$/d;}' > /backup/$(date '+%Y-%m-%d')/$table.data.sql;
# done;

