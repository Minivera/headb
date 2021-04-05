module.exports = {
  development: {
    client: 'pg',
    connection: process.env.DB_URL || 'postgresql://headb:headb@localhost:5432/head_db',
    migrations: {
      tableName: 'knex_migrations',
      directory: 'migrations',
    },
  },
};
