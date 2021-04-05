import feathers from '@feathersjs/feathers';
import '@feathersjs/transport-commons';
import express from '@feathersjs/express';

import { db } from './db';
import { UserService } from './services/userService';
import { CollectionService } from './services/collectionService';
import { DocumentService } from './services/documentService';

const app = express(feathers());

app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.configure(express.rest());

app.use(
  '/users',
  new UserService({
    Model: db,
    name: 'users',
  })
);

app.use(
  '/users/:userId/collections',
  new CollectionService(
    {
      Model: db,
      name: 'collections',
    },
    app
  )
);

app.use(
  '/users/:userId/collections/:collectionId/documents',
  new DocumentService(
    {
      Model: db,
      name: 'documents',
    },
    app
  )
);

app.use(express.errorHandler());

// Start the server
app
  .listen(process.env.PORT || 3000)
  .on('listening', () => console.log(`Server listening on localhost:${process.env.PORT || 3000}`));

export { app };
