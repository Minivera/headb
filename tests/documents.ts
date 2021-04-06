import { suite } from '@minivera/testifyjs';
import fetch from 'node-fetch';
import expect from 'expect';

import { User } from '../src/services/userService';
import { Collection } from '../src/services/collectionService';
import { Document } from '../src/services/documentService';
import {
  cleanUsers,
  createUser,
  cleanCollections,
  createCollection,
  cleanDocuments,
  createDocuments,
  createDocument,
  fetchDocuments,
  fetchDocument,
  updateDocument,
  deleteDocument,
  CreateUserOutput,
  CreateCollectionOutput,
  DocumentsOutput,
  CreateDocumentsOutput,
  DocumentInput,
  CreateDocumentOutput,
  UpdateDocumentInput,
  UpdateDocumentOutput,
  DeleteDocumentOutput,
  DocumentOutput,
} from './utils';

const port = process.env.PORT || 3000;

interface ErrorMessage {
  name: string;
  message: string;
  code: number;
}

suite('The documents service', suite => {
  let user: User;
  let collection: Collection;

  suite.beforeEach(async () => {
    const result1 = (await createUser({})) as CreateUserOutput;
    user = result1.created;

    const result2 = (await createCollection({ userId: user.id as string })) as CreateCollectionOutput;
    collection = result2.created;
  });

  suite.afterEach(async () => {
    const documents: Document[] = await fetch(
      `http://localhost:${port}/users/${user.id}/collections/${collection.id}/documents`
    ).then(response => response.json());
    await cleanDocuments({ userId: user.id as string, collectionId: collection.id as string, documents });
    await cleanCollections({ userId: user.id as string, collections: [collection] });
    await cleanUsers({ users: [user] });
  });

  suite.suite('The GET documents endpoint', suite => {
    suite.test('Will return an empty array when there are no documents', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id }))
        .act(fetchDocuments)
        .assert<DocumentsOutput>(({ documents }) => {
          expect(documents).toEqual([]);
        });
    });

    suite.test('Will return an array with existing documents', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id }))
        .arrange(createDocuments)
        .act(fetchDocuments)
        .assert<DocumentsOutput & CreateDocumentsOutput>(({ created, documents }) => {
          expect(documents).toEqual(expect.arrayContaining(created));
        });
    });

    suite.test('Will trigger a 404 if the collection cannot be found', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchDocuments)
        .assert<{ documents: ErrorMessage }>(({ documents }) => {
          expect(documents.code).toBe(404);
        });
    });
  });

  suite.suite('The GET document by id endpoint', suite => {
    suite.afterEach(async () => {
      const documents: Document[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections/${collection.id}/documents`
      ).then(response => response.json());
      await cleanDocuments({ userId: user.id as string, collectionId: collection.id as string, documents });
    });

    suite.test('Will return a 404 when the document cannot be found', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id, id: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchDocument)
        .assert<{ document: ErrorMessage }>(({ document }) => {
          expect(document.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the collection does not exist', test => {
      test
        .arrange(() => ({
          userId: user.id,
          collectionId: '8aec7349-5d4f-4dce-b576-182841348a3e',
          id: '8aec7349-5d4f-4dce-b576-182841348a3e',
        }))
        .act(fetchDocument)
        .assert<{ document: ErrorMessage }>(({ document }) => {
          expect(document.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the document id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id, id: '1234' }))
        .act(fetchDocument)
        .assert<{ document: ErrorMessage }>(({ document }) => {
          expect(document.code).toBe(400);
        });
    });

    suite.test('Will return a 400 when the collection id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: '1234', id: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchDocument)
        .assert<{ document: ErrorMessage }>(({ document }) => {
          expect(document.code).toBe(400);
        });
    });

    suite.test('Will return the document object when there is an existing document', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id }))
        .arrange(createDocument)
        .arrange<CreateDocumentOutput, CreateDocumentOutput & DocumentInput>(({ created }) => ({
          userId: user.id as string,
          collectionId: collection.id as string,
          id: created.id,
          created,
        }))
        .act(fetchDocument)
        .assert<DocumentOutput & CreateDocumentOutput>(({ created, document }) => {
          expect(document).toEqual(created);
        });
    });
  });

  suite.suite('The POST document endpoint', suite => {
    suite.afterEach(async () => {
      const documents: Document[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections/${collection.id}/documents`
      ).then(response => response.json());
      await cleanDocuments({ userId: user.id as string, collectionId: collection.id as string, documents });
    });

    suite.test('Will return the document object once created', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id }))
        .act(createDocument)
        .assert<CreateDocumentOutput>(({ created }) => {
          expect(created).toBeDefined();
        });
    });

    suite.test('Will trigger a 404 if the collection cannot be found', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(createDocument)
        .assert<{ created: ErrorMessage }>(({ created }) => {
          expect(created.code).toBe(404);
        });
    });

    suite.test('Will trigger a 400 if the collection id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: '1234' }))
        .act(createDocument)
        .assert<{ created: ErrorMessage }>(({ created }) => {
          expect(created.code).toBe(400);
        });
    });

    suite.test('Will trigger a 400 if the values are invalid', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id, content: null }))
        .act(createDocument)
        .assert<{ created: ErrorMessage }>(({ created }) => {
          expect(created.code).toBe(400);
        });
    });
  });

  suite.suite('The PATCH document by id endpoint', suite => {
    suite.afterEach(async () => {
      const documents: Document[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections/${collection.id}/documents`
      ).then(response => response.json());
      await cleanDocuments({ userId: user.id as string, collectionId: collection.id as string, documents });
    });

    suite.test('Will return a 404 when there are no documents', test => {
      test
        .arrange(() => ({
          userId: user.id,
          collectionId: '8aec7349-5d4f-4dce-b576-182841348a3e',
          document: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', content: {} },
        }))
        .act(updateDocument)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the document does not exist', test => {
      test
        .arrange(() => ({
          userId: user.id,
          collectionId: collection.id,
          document: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', content: {} },
        }))
        .act(updateDocument)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the document id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id, document: { id: '1234', content: {} } }))
        .act(updateDocument)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(400);
        });
    });

    suite.test('Will return a 400 when the collection id is malformed', test => {
      test
        .arrange(() => ({
          userId: user.id,
          collectionId: '1234',
          document: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', content: {} },
        }))
        .act(updateDocument)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(400);
        });
    });

    suite.test('Will return the updated document object when there is an existing document', test => {
      const newContent = { foo: 'bar' };

      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id }))
        .arrange(createDocument)
        .arrange<CreateDocumentOutput, CreateDocumentOutput & UpdateDocumentInput>(({ created }) => ({
          userId: user.id as string,
          collectionId: collection.id as string,
          document: { ...created, content: newContent },
          created,
        }))
        .act(updateDocument)
        .assert<CreateDocumentOutput & UpdateDocumentOutput>(({ created, updated }) => {
          expect(updated.id).toBe(created.id);
          expect(updated.content).toEqual(newContent);
          expect(updated.createdAt).toBe(created.createdAt);
        });
    });
  });

  suite.suite('The DELETE document by id endpoint', suite => {
    suite.afterEach(async () => {
      const documents: Document[] = await fetch(
        `http://localhost:${port}/users/${user.id}/collections/${collection.id}/documents`
      ).then(response => response.json());
      await cleanDocuments({ userId: user.id as string, collectionId: collection.id as string, documents });
    });

    suite.test('Will return a 404 when there are no documents', test => {
      test
        .arrange(() => ({
          userId: user.id,
          collectionId: collection.id,
          document: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', content: {} },
        }))
        .act(deleteDocument)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the collection does not exist', test => {
      test
        .arrange(() => ({
          userId: user.id,
          collectionId: '8aec7349-5d4f-4dce-b576-182841348a3e',
          document: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', content: {} },
        }))
        .act(deleteDocument)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the document id is malformed', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id, document: { id: '1234', content: {} } }))
        .act(deleteDocument)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(400);
        });
    });

    suite.test('Will return a 400 when the collection id is malformed', test => {
      test
        .arrange(() => ({
          userId: user.id,
          collectionId: '1234',
          document: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', content: {} },
        }))
        .act(deleteDocument)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(400);
        });
    });

    suite.test('Will return the deleted document object when there is an existing document', test => {
      test
        .arrange(() => ({ userId: user.id, collectionId: collection.id }))
        .arrange(createDocument)
        .arrange<CreateDocumentOutput, CreateDocumentOutput & UpdateDocumentInput>(({ created }) => ({
          userId: user.id as string,
          collectionId: collection.id as string,
          document: created,
          created,
        }))
        .act(deleteDocument)
        .act(fetchDocuments)
        .assert<CreateDocumentOutput & DeleteDocumentOutput & DocumentsOutput>(({ created, deleted, documents }) => {
          expect(deleted).toEqual(created);

          expect(documents).not.toContain(deleted);
        });
    });
  });
});
