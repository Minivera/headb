import { KnexServiceOptions, Service } from 'feathers-knex';
import { Application, Id, NullableId, Params } from '@feathersjs/feathers';
import { BadRequest, NotFound } from '@feathersjs/errors';

import { CollectionService } from './collectionService';
import { validate } from 'uuid';

export interface Document {
  id?: string;
  content: unknown;
  // eslint-disable-next-line camelcase
  collection_id: string;
  updatedAt?: Date;
  createdAt?: Date;
}

export class DocumentService extends Service<Document> {
  private app: Application;

  constructor(options: Partial<KnexServiceOptions>, app: Application) {
    super(options);
    this.app = app;
  }

  async find(params: Params): Promise<Document[]> {
    const collectionId = params.route?.collectionId;

    return this.knex('documents')
      .where({
        collection_id: await this.getCollectionId(collectionId, params),
      })
      .select();
  }

  async get(id: Id, params: Params): Promise<Document> {
    if (!validate(id.toString())) {
      throw new BadRequest(`Invalid uuid ${id}`);
    }

    const collectionId = params.route?.collectionId;

    const found = await this.knex('documents')
      .where({
        id: id,
        collection_id: await this.getCollectionId(collectionId, params),
      })
      .select()
      .first();

    if (!found) {
      throw new NotFound(`Could not find document for id ${id}`);
    }

    return found;
  }

  async create(data: Document, params: Params): Promise<Document | Document[]> {
    const collectionId = params.route?.collectionId;

    return super.create(
      {
        ...data,
        collection_id: await this.getCollectionId(collectionId, params),
      },
      params
    );
  }

  async update(id: NullableId, data: Document, params: Params): Promise<Document> {
    const collectionId = params.route?.collectionId;
    await this.getCollectionId(collectionId, params);

    if (!id) {
      throw new BadRequest('You need to provide an ID to update');
    }

    const prev = await this.get(id, params);

    return super.update(
      id,
      {
        ...prev,
        ...data,
      } as Document,
      params
    );
  }

  async patch(id: NullableId, data: Document, params: Params): Promise<Document> {
    return this.update(id, data, params);
  }

  async remove(id: NullableId, params: Params): Promise<Document | Document[]> {
    const collectionId = params.route?.collectionId;
    await this.getCollectionId(collectionId, params);

    if (!id) {
      throw new BadRequest('You need to provide an ID to update');
    }

    return super.remove(id, params);
  }

  private async getCollectionId(collectionId: string | undefined, params: Params) {
    if (!collectionId) {
      throw new BadRequest('You need to provide a collectionID.');
    }

    if (!validate(collectionId)) {
      throw new BadRequest(`Invalid uuid ${collectionId}`);
    }

    const collectionService = this.app.service('users/:userId/collections') as CollectionService;
    let foundCollectionId;
    try {
      const collection = await collectionService.get(collectionId, params);

      if (!collection) {
        throw new Error('catch');
      }

      foundCollectionId = collection.id;
    } catch {
      throw new NotFound(`Could not find collection for id ${collectionId}`);
    }
    return foundCollectionId;
  }
}
