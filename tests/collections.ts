import { suite } from '@minivera/testifyjs';
import fetch from 'node-fetch';
import expect from 'expect';

import { User } from '../src/services/userService';
import { Collection } from '../src/services/collectionService';
import {
  cleanUsers,
  createUser,
  cleanCollections,
  createCollections,
  createCollection,
  fetchCollections,
  fetchCollection,
  updateCollection,
  deleteCollection,
  CreateUserOutput,
  CollectionsOutput,
  CollectionInput,
  CreateCollectionsOutput,
  CreateCollectionOutput,
  CollectionOutput,
  UpdateCollectionInput,
  UpdateCollectionOutput,
  DeleteCollectionOutput,
} from './utils';

const port = process.env.PORT || 3000;

interface ErrorMessage {
  name: string;
  message: string;
  code: number;
}

suite('The collections service', suite => {
  let user: User;

  suite.beforeEach(async () => {
    const result = (await createUser({})) as CreateUserOutput;
    user = result.created;
  });

  suite.afterEach(async () => {
    const collections: Collection[] = await fetch(
      `http://localhost:${port}/users/${user.id}/collections`
    ).then(response => response.json());
    await cleanCollections({ userId: user.id as string, collections });
    await cleanUsers({ users: [user] });
  });

  suite.suite('The GET collections endpoint', suite => {
    suite.test('Will return an empty array when there are no collections', test => {
      test
        .arrange(() => ({ userId: user.id }))
        .act(fetchCollections)
        .assert<CollectionsOutput>(({ collections }) => {
          expect(collections).toEqual([]);
        });
    });

    suite.test('Will return an array with existing collections', test => {
      test
        .arrange(() => ({ userId: user.id }))
        .arrange(createCollections)
        .act(fetchCollections)
        .assert<CollectionsOutput & CreateCollectionsOutput>(({ created, collections }) => {
          expect(collections).toEqual(expect.arrayContaining(created));
        });
    });

    suite.test('Will trigger a 404 if the user cannot be found', test => {
      test
        .arrange(() => ({ userId: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchCollections)
        .assert<{ collections: ErrorMessage }>(({ collections }) => {
          expect(collections.code).toBe(404);
        });
    });
  });

  suite.suite('The GET collection by id endpoint', suite => {
    suite.afterEach(async () => {
      const collections: Collection[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections`
      ).then(response => response.json());
      await cleanCollections({ userId: user.id as string, collections });
    });

    suite.test('Will return a 404 when the collection cannot be found', test => {
      test
        .arrange(() => ({ userId: user.id, id: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchCollection)
        .assert<{ collection: ErrorMessage }>(({ collection }) => {
          expect(collection.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the user does not exist', test => {
      test
        .arrange(() => ({ userId: '8aec7349-5d4f-4dce-b576-182841348a3e', id: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchCollection)
        .assert<{ collection: ErrorMessage }>(({ collection }) => {
          expect(collection.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the collection id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, id: '1234' }))
        .act(fetchCollection)
        .assert<{ collection: ErrorMessage }>(({ collection }) => {
          expect(collection.code).toBe(400);
        });
    });

    suite.test('Will return a 400 when the user id is malformed', test => {
      test
        .arrange(() => ({ userId: '1234', id: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchCollection)
        .assert<{ collection: ErrorMessage }>(({ collection }) => {
          expect(collection.code).toBe(400);
        });
    });

    suite.test('Will return the collection object when there is an existing collection', test => {
      test
        .arrange(() => ({ userId: user.id }))
        .arrange(createCollection)
        .arrange<CreateCollectionOutput, CreateCollectionOutput & CollectionInput>(({ created }) => ({
          userId: user.id as string,
          id: created.id,
          created,
        }))
        .act(fetchCollection)
        .assert<CollectionOutput & CreateCollectionOutput>(({ created, collection }) => {
          expect(collection).toEqual(created);
        });
    });
  });

  suite.suite('The POST collection endpoint', suite => {
    suite.afterEach(async () => {
      const collections: Collection[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections`
      ).then(response => response.json());
      await cleanCollections({ userId: user.id as string, collections });
    });

    suite.test('Will return the collection object once created', test => {
      test
        .arrange(() => ({ userId: user.id }))
        .act(createCollection)
        .assert<CreateCollectionOutput>(({ created }) => {
          expect(created).toBeDefined();
        });
    });

    suite.test('Will trigger a 404 if the user cannot be found', test => {
      test
        .arrange(() => ({ userId: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(createCollection)
        .assert<{ created: ErrorMessage }>(({ created }) => {
          expect(created.code).toBe(404);
        });
    });

    suite.test('Will trigger a 400 if the user id is malformed', test => {
      test
        .arrange(() => ({ userId: '1234' }))
        .act(createCollection)
        .assert<{ created: ErrorMessage }>(({ created }) => {
          expect(created.code).toBe(400);
        });
    });

    suite.test('Will trigger a 400 if the values are invalid', test => {
      test
        .arrange(() => ({ userId: user.id, name: null }))
        .act(createCollection)
        .assert<{ created: ErrorMessage }>(({ created }) => {
          expect(created.code).toBe(400);
        });
    });
  });

  suite.suite('The PATCH collection by id endpoint', suite => {
    suite.afterEach(async () => {
      const collections: Collection[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections`
      ).then(response => response.json());
      await cleanCollections({ userId: user.id as string, collections });
    });

    suite.test('Will return a 404 when there are no collections', test => {
      test
        .arrange(() => ({ userId: user.id, collection: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', name: 'test' } }))
        .act(updateCollection)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the user does not exist', test => {
      test
        .arrange(() => ({
          userId: '8aec7349-5d4f-4dce-b576-182841348a3e',
          collection: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', name: 'test' },
        }))
        .act(updateCollection)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the collection id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, collection: { id: '1234', name: 'test' } }))
        .act(updateCollection)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(400);
        });
    });

    suite.test('Will return a 400 when the user id is malformed', test => {
      test
        .arrange(() => ({ userId: '1234', collection: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', name: 'test' } }))
        .act(updateCollection)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(400);
        });
    });

    suite.test('Will return the updated collection object when there is an existing collection', test => {
      const newName = 'foo';

      test
        .arrange(() => ({ userId: user.id }))
        .arrange(createCollection)
        .arrange<CreateCollectionOutput, CreateCollectionOutput & UpdateCollectionInput>(({ created }) => ({
          userId: user.id as string,
          collection: { ...created, name: newName },
          created,
        }))
        .act(updateCollection)
        .assert<CreateCollectionOutput & UpdateCollectionOutput>(({ created, updated }) => {
          expect(updated.id).toBe(created.id);
          expect(updated.name).toBe(newName);
          expect(updated.createdAt).toBe(created.createdAt);
        });
    });
  });

  suite.suite('The DELETE collection by id endpoint', suite => {
    suite.afterEach(async () => {
      const collections: Collection[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections`
      ).then(response => response.json());
      await cleanCollections({ userId: user.id as string, collections });
    });

    suite.test('Will return a 404 when there are no collections', test => {
      test
        .arrange(() => ({ userId: user.id, collection: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', name: 'test' } }))
        .act(deleteCollection)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the user does not exist', test => {
      test
        .arrange(() => ({
          userId: '8aec7349-5d4f-4dce-b576-182841348a3e',
          collection: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', name: 'test' },
        }))
        .act(deleteCollection)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the collection id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, collection: { id: '1234', name: 'test' } }))
        .act(deleteCollection)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(400);
        });
    });

    suite.test('Will return a 400 when the user id is malformed', test => {
      test
        .arrange(() => ({ userId: '1234', collection: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', name: 'test' } }))
        .act(deleteCollection)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(400);
        });
    });

    suite.test('Will return the deleted collection object when there is an existing collection', test => {
      test
        .arrange(() => ({ userId: user.id }))
        .arrange(createCollection)
        .arrange<CreateCollectionOutput, CreateCollectionOutput & UpdateCollectionInput>(({ created }) => ({
          userId: user.id as string,
          collection: created,
          created,
        }))
        .act(deleteCollection)
        .act(fetchCollections)
        .assert<CreateCollectionOutput & DeleteCollectionOutput & CollectionsOutput>(
          ({ created, deleted, collections }) => {
            expect(deleted).toEqual(created);

            expect(collections).not.toContain(deleted);
          }
        );
    });
  });
});
