import { ExecutionFunction, RunnerParams } from '@minivera/testifyjs';
import fetch from 'node-fetch';

import { User } from '../src/services/userService';
import { Collection } from '../src/services/collectionService';
import { Document } from '../src/services/documentService';

const port = process.env.PORT || 3000;

export interface UsersOutput {
  users: User[];
}

export const fetchUsers: ExecutionFunction<RunnerParams, UsersOutput> = async (params): Promise<UsersOutput> => {
  const users = await fetch(`http://localhost:${port}/users`).then(response => response.json());

  return {
    users,
    ...params,
  };
};

export interface FetchUserInput {
  id: string;
}

export interface UserOutput {
  user: User;
}

export const fetchUser: ExecutionFunction<FetchUserInput, FetchUserInput & UserOutput> = async ({
  id,
  ...rest
}): Promise<FetchUserInput & UserOutput> => {
  const user = await fetch(`http://localhost:${port}/users/${id}`).then(response => response.json());

  return {
    user,
    id,
    ...rest,
  };
};

export interface UserInput {
  username?: string;
}

export interface CreateUserOutput {
  created: User;
}

export const createUser: ExecutionFunction<UserInput, UserInput & CreateUserOutput> = async ({
  username = 'test',
  ...rest
}): Promise<UserInput & CreateUserOutput> => {
  const created = await fetch(`http://localhost:${port}/users`, {
    method: 'POST',
    body: JSON.stringify({
      username,
    }),
    headers: { 'Content-Type': 'application/json' },
  }).then(response => response.json());

  return {
    created,
    username,
    ...rest,
  };
};

export interface UsersInput {
  count?: number;
}

export interface CreateUsersOutput {
  created: User[];
}

export const createUsers: ExecutionFunction<UsersInput, UsersInput & CreateUsersOutput> = async ({
  count = 5,
  ...rest
}): Promise<UsersInput & CreateUsersOutput> => {
  const promises: Promise<User>[] = [];

  for (let i = 0; i < count; i++) {
    promises.push(
      fetch(`http://localhost:${port}/users`, {
        method: 'POST',
        body: JSON.stringify({
          username: `${i}`,
        }),
        headers: { 'Content-Type': 'application/json' },
      }).then(response => response.json())
    );
  }

  const created = await Promise.all(promises);

  return {
    created,
    count,
    ...rest,
  };
};

export interface UpdateUserOutput {
  updated: User;
}

export const updateUser: ExecutionFunction<UserOutput, UserOutput & UpdateUserOutput> = async ({
  user,
  ...rest
}): Promise<UserOutput & UpdateUserOutput> => {
  const updated = await fetch(`http://localhost:${port}/users/${user.id}`, {
    method: 'PATCH',
    body: JSON.stringify({
      username: user.username,
    }),
    headers: { 'Content-Type': 'application/json' },
  }).then(response => response.json());

  return {
    updated,
    user,
    ...rest,
  };
};

export interface DeleteUserOutput {
  deleted: User;
}

export const deleteUser: ExecutionFunction<UserOutput, UserOutput & DeleteUserOutput> = async ({
  user,
  ...rest
}): Promise<UserOutput & DeleteUserOutput> => {
  const deleted = await fetch(`http://localhost:${port}/users/${user.id}`, {
    method: 'DELETE',
  }).then(response => response.json());

  return {
    deleted,
    user,
    ...rest,
  };
};

export const cleanUsers: ExecutionFunction<UsersOutput, UsersOutput> = async ({ users }): Promise<void> => {
  await Promise.all(
    users.map(user =>
      fetch(`http://localhost:${port}/users/${user.id}`, {
        method: 'DELETE',
      })
    )
  );
};

export interface CollectionPathInput {
  userId: string;
}

export interface CollectionsOutput extends CollectionPathInput {
  collections: Collection[];
}

export const fetchCollections: ExecutionFunction<CollectionPathInput, CollectionsOutput> = async ({
  userId,
  ...rest
}): Promise<CollectionsOutput> => {
  const collections = await fetch(`http://localhost:${port}/users/${userId}/collections`).then(response =>
    response.json()
  );

  return {
    collections,
    userId,
    ...rest,
  };
};

export interface FetchCollectionInput extends CollectionPathInput {
  id: string;
}

export interface CollectionOutput {
  collection: Collection;
}

export const fetchCollection: ExecutionFunction<
  FetchCollectionInput,
  FetchCollectionInput & CollectionOutput
> = async ({ userId, id, ...rest }): Promise<FetchCollectionInput & CollectionOutput> => {
  const collection = await fetch(`http://localhost:${port}/users/${userId}/collections/${id}`).then(response =>
    response.json()
  );

  return {
    collection,
    id,
    userId,
    ...rest,
  };
};

export interface CollectionInput extends CollectionPathInput {
  name?: string;
}

export interface CreateCollectionOutput {
  created: Collection;
}

export const createCollection: ExecutionFunction<CollectionInput, CollectionInput & CreateCollectionOutput> = async ({
  name = 'test',
  userId,
  ...rest
}): Promise<CollectionInput & CreateCollectionOutput> => {
  const created = await fetch(`http://localhost:${port}/users/${userId}/collections`, {
    method: 'POST',
    body: JSON.stringify({
      name,
    }),
    headers: { 'Content-Type': 'application/json' },
  }).then(response => response.json());

  return {
    created,
    name,
    userId,
    ...rest,
  };
};

export interface CollectionsInput extends CollectionPathInput {
  count?: number;
}

export interface CreateCollectionsOutput {
  created: Collection[];
}

export const createCollections: ExecutionFunction<
  CollectionsInput,
  CollectionsInput & CreateCollectionsOutput
> = async ({ count = 5, userId, ...rest }): Promise<CollectionsInput & CreateCollectionsOutput> => {
  const promises: Promise<Collection>[] = [];

  for (let i = 0; i < count; i++) {
    promises.push(
      fetch(`http://localhost:${port}/users/${userId}/collections`, {
        method: 'POST',
        body: JSON.stringify({
          name: `${i}`,
        }),
        headers: { 'Content-Type': 'application/json' },
      }).then(response => response.json())
    );
  }

  const created = await Promise.all(promises);

  return {
    created,
    userId,
    count,
    ...rest,
  };
};

export interface UpdateCollectionInput extends CollectionPathInput {
  collection: Collection;
}

export interface UpdateCollectionOutput {
  updated: Collection;
}

export const updateCollection: ExecutionFunction<
  UpdateCollectionInput,
  UpdateCollectionInput & UpdateCollectionOutput
> = async ({ collection, userId, ...rest }): Promise<UpdateCollectionInput & UpdateCollectionOutput> => {
  const updated = await fetch(`http://localhost:${port}/users/${userId}/collections/${collection.id}`, {
    method: 'PATCH',
    body: JSON.stringify({
      name: collection.name,
    }),
    headers: { 'Content-Type': 'application/json' },
  }).then(response => response.json());

  return {
    updated,
    collection,
    userId,
    ...rest,
  };
};

export interface DeleteCollectionOutput {
  deleted: Collection;
}

export const deleteCollection: ExecutionFunction<
  UpdateCollectionInput,
  UpdateCollectionInput & DeleteCollectionOutput
> = async ({ collection, userId, ...rest }): Promise<UpdateCollectionInput & DeleteCollectionOutput> => {
  const deleted = await fetch(`http://localhost:${port}/users/${userId}/collections/${collection.id}`, {
    method: 'DELETE',
  }).then(response => response.json());

  return {
    deleted,
    collection,
    userId,
    ...rest,
  };
};

export const cleanCollections: ExecutionFunction<CollectionsOutput, CollectionsOutput> = async ({
  collections,
  userId,
}): Promise<void> => {
  await Promise.all(
    collections.map(collection =>
      fetch(`http://localhost:${port}/users/${userId}/collections/${collection.id}`, {
        method: 'DELETE',
      })
    )
  );
};

export interface DocumentPathInput {
  userId: string;
  collectionId: string;
}

export interface DocumentsOutput extends DocumentPathInput {
  documents: Document[];
}

export const fetchDocuments: ExecutionFunction<DocumentPathInput, DocumentsOutput> = async ({
  userId,
  collectionId,
  ...rest
}): Promise<DocumentsOutput> => {
  const documents = await fetch(
    `http://localhost:${port}/users/${userId}/collections/${collectionId}/documents`
  ).then(response => response.json());

  return {
    documents,
    userId,
    collectionId,
    ...rest,
  };
};

export interface FetchDocumentInput extends DocumentPathInput {
  id: string;
}

export interface DocumentOutput {
  document: Document;
}

export const fetchDocument: ExecutionFunction<FetchDocumentInput, FetchDocumentInput & DocumentOutput> = async ({
  userId,
  collectionId,
  id,
  ...rest
}): Promise<FetchDocumentInput & DocumentOutput> => {
  const document = await fetch(
    `http://localhost:${port}/users/${userId}/collections/${collectionId}/documents/${id}`
  ).then(response => response.json());

  return {
    document,
    id,
    userId,
    collectionId,
    ...rest,
  };
};

export interface DocumentInput extends DocumentPathInput {
  content?: unknown;
}

export interface CreateDocumentOutput {
  created: Document;
}

export const createDocument: ExecutionFunction<DocumentInput, DocumentInput & CreateDocumentOutput> = async ({
  content = {},
  userId,
  collectionId,
  ...rest
}): Promise<DocumentInput & CreateDocumentOutput> => {
  const created = await fetch(`http://localhost:${port}/users/${userId}/collections/${collectionId}/documents`, {
    method: 'POST',
    body: JSON.stringify({
      content,
    }),
    headers: { 'Content-Type': 'application/json' },
  }).then(response => response.json());

  return {
    created,
    content,
    userId,
    collectionId,
    ...rest,
  };
};

export interface DocumentsInput extends DocumentPathInput {
  count?: number;
}

export interface CreateDocumentsOutput {
  created: Document[];
}

export const createDocuments: ExecutionFunction<DocumentsInput, DocumentsInput & CreateDocumentsOutput> = async ({
  count = 5,
  userId,
  collectionId,
  ...rest
}): Promise<DocumentsInput & CreateDocumentsOutput> => {
  const promises: Promise<Document>[] = [];

  for (let i = 0; i < count; i++) {
    promises.push(
      fetch(`http://localhost:${port}/users/${userId}/collections/${collectionId}/documents`, {
        method: 'POST',
        body: JSON.stringify({
          content: {
            count: i,
          },
        }),
        headers: { 'Content-Type': 'application/json' },
      }).then(response => response.json())
    );
  }

  const created = await Promise.all(promises);

  return {
    created,
    userId,
    collectionId,
    count,
    ...rest,
  };
};

export interface UpdateDocumentInput extends DocumentPathInput {
  document: Document;
}

export interface UpdateDocumentOutput {
  updated: Document;
}

export const updateDocument: ExecutionFunction<
  UpdateDocumentInput,
  UpdateDocumentInput & UpdateDocumentOutput
> = async ({ document, userId, collectionId, ...rest }): Promise<UpdateDocumentInput & UpdateDocumentOutput> => {
  const updated = await fetch(
    `http://localhost:${port}/users/${userId}/collections/${collectionId}/documents/${document.id}`,
    {
      method: 'PATCH',
      body: JSON.stringify({
        content: document.content,
      }),
      headers: { 'Content-Type': 'application/json' },
    }
  ).then(response => response.json());

  return {
    updated,
    document,
    userId,
    collectionId,
    ...rest,
  };
};

export interface DeleteDocumentOutput {
  deleted: Document;
}

export const deleteDocument: ExecutionFunction<
  UpdateDocumentInput,
  UpdateDocumentInput & DeleteDocumentOutput
> = async ({ document, userId, collectionId, ...rest }): Promise<UpdateDocumentInput & DeleteDocumentOutput> => {
  const deleted = await fetch(
    `http://localhost:${port}/users/${userId}/collections/${collectionId}/documents/${document.id}`,
    {
      method: 'DELETE',
    }
  ).then(response => response.json());

  return {
    deleted,
    document,
    userId,
    collectionId,
    ...rest,
  };
};

export const cleanDocuments: ExecutionFunction<DocumentsOutput, DocumentsOutput> = async ({
  documents,
  collectionId,
  userId,
}): Promise<void> => {
  await Promise.all(
    documents.map(document =>
      fetch(`http://localhost:${port}/users/${userId}/collections/${collectionId}/documents/${document.id}`, {
        method: 'DELETE',
      })
    )
  );
};
