import knex from 'knex';

export const db = knex({
  client: 'pg',
  connection: process.env.DB_URL || 'postgresql://headb:headb@localhost:5432/head_db',
  migrations: {
    tableName: 'knex_migrations',
    extension: 'ts',
    directory: 'migrations',
  },
});
