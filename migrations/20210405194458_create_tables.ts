import { Knex } from 'knex';

export async function up(knex: Knex): Promise<void> {
  return knex.schema
    .createTable('users', table => {
      table.uuid('id').primary().defaultTo(knex.raw('(gen_random_uuid())'));
      table.string('username').unique().index().notNullable();
      table.timestamp('createdAt').defaultTo(knex.fn.now());
      table.timestamp('updatedAt').defaultTo(knex.fn.now());
    })
    .createTable('collections', table => {
      table.uuid('id').primary().defaultTo(knex.raw('(gen_random_uuid())'));
      table.string('name').notNullable();
      table.uuid('user_id').index().notNullable().references('id').inTable('users').onDelete('CASCADE');
      table.timestamp('createdAt').defaultTo(knex.fn.now());
      table.timestamp('updatedAt').defaultTo(knex.fn.now());
    })
    .createTable('documents', table => {
      table.uuid('id').primary().defaultTo(knex.raw('(gen_random_uuid())'));
      table.json('content').notNullable().defaultTo('{}');
      table.uuid('collection_id').index().notNullable().references('id').inTable('collections').onDelete('CASCADE');
      table.timestamp('createdAt').defaultTo(knex.fn.now());
      table.timestamp('updatedAt').defaultTo(knex.fn.now());
    });
}

export async function down(knex: Knex): Promise<void> {
  return knex.schema.dropTable('documents').dropTable('collections').dropTable('users');
}
