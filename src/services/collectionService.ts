import { Application, Id, NullableId, Params } from '@feathersjs/feathers';
import { Service, KnexServiceOptions } from 'feathers-knex';
import { BadRequest, NotFound } from '@feathersjs/errors';
import { validate } from 'uuid';

import { UserService } from './userService';

export interface Collection {
  id?: string;
  name: string;
  // eslint-disable-next-line camelcase
  user_id: string;
  updatedAt?: Date;
  createdAt?: Date;
}

export class CollectionService extends Service<Collection> {
  private app: Application;

  constructor(options: Partial<KnexServiceOptions>, app: Application) {
    super(options);
    this.app = app;
  }

  async find(params: Params): Promise<Collection[]> {
    const userId = params.route?.userId;

    return this.knex('collections')
      .where({
        user_id: await this.getUserId(userId),
      })
      .select();
  }

  async get(id: Id, params: Params): Promise<Collection> {
    const userId = params.route?.userId;

    if (!validate(id.toString())) {
      throw new BadRequest(`Invalid uuid ${id}`);
    }

    const found = await this.knex('collections')
      .where({
        id: id,
        user_id: await this.getUserId(userId),
      })
      .select()
      .first();

    if (!found) {
      throw new NotFound(`Could not find collection for id ${id}`);
    }

    return found;
  }

  async create(data: Collection, params: Params): Promise<Collection | Collection[]> {
    const userId = params.route?.userId;

    return super.create(
      {
        ...data,
        user_id: await this.getUserId(userId),
      },
      params
    );
  }

  async update(id: NullableId, data: Collection, params: Params): Promise<Collection> {
    const userId = params.route?.userId;
    await this.getUserId(userId);

    if (!id) {
      throw new BadRequest('You need to provide an ID to update');
    }

    const prev = await this.get(id, params);

    return super.update(
      id,
      {
        ...prev,
        ...data,
      } as Collection,
      params
    );
  }

  async patch(id: NullableId, data: Collection, params: Params): Promise<Collection> {
    return this.update(id, data, params);
  }

  async remove(id: NullableId, params: Params): Promise<Collection | Collection[]> {
    const userId = params.route?.userId;
    await this.getUserId(userId);

    if (!id) {
      throw new BadRequest('You need to provide an ID to delete');
    }

    return super.remove(id, params);
  }

  private async getUserId(userId?: string) {
    if (!userId) {
      throw new BadRequest('You need to provide a userID.');
    }

    if (!validate(userId)) {
      throw new BadRequest(`Invalid uuid ${userId}`);
    }

    const userService = this.app.service('users') as UserService;
    let foundUserId;
    try {
      const user = await userService.get(userId);

      if (!user) {
        throw new Error('catch');
      }

      foundUserId = user.id;
    } catch {
      throw new NotFound(`Could not find user for id ${userId}`);
    }
    return foundUserId;
  }
}
