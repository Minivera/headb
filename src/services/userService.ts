import { Service } from 'feathers-knex';

export interface User {
  id?: string;
  username: string;
  updatedAt: Date;
  createdAt: Date;
}

export class UserService extends Service<User> {}
