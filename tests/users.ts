import { suite } from '@minivera/testifyjs';
import fetch from 'node-fetch';
import expect from 'expect';

import { User } from '../src/services/userService';
import {
  cleanUsers,
  createUsers,
  createUser,
  fetchUsers,
  fetchUser,
  updateUser,
  deleteUser,
  UsersOutput,
  CreateUsersOutput,
  UserOutput,
  CreateUserOutput,
  UserInput,
  UpdateUserOutput,
  DeleteUserOutput,
} from './utils';

const port = process.env.PORT || 3000;

interface ErrorMessage {
  name: string;
  message: string;
  code: number;
}

suite('The users service', suite => {
  suite.afterEach(async () => {
    const users: User[] = await fetch(`http://localhost:${port}/users`).then(response => response.json());
    await cleanUsers({ users });
  });

  suite.suite('The GET users endpoint', suite => {
    suite.test('Will return an empty array when there are no users', test => {
      test.act(fetchUsers).assert<UsersOutput>(({ users }) => {
        expect(users).toEqual([]);
      });
    });

    suite.test('Will return an array with existing users', test => {
      test
        .arrange(createUsers)
        .act(fetchUsers)
        .assert<UsersOutput & CreateUsersOutput>(({ created, users }) => {
          expect(users).toEqual(created);
        });
    });
  });

  suite.suite('The GET user by id endpoint', suite => {
    suite.afterEach(async () => {
      const users: User[] = await fetch(`http://localhost:${port}/users`).then(response => response.json());
      await cleanUsers({ users });
    });

    suite.test('Will return a 404 when there are no users', test => {
      test
        .arrange(() => ({ id: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchUser)
        .assert<{ user: ErrorMessage }>(({ user }) => {
          expect(user.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the user does not exist', test => {
      test
        .arrange(createUser)
        .arrange(() => ({ id: '8aec7349-5d4f-4dce-b576-182841348a3e' }))
        .act(fetchUser)
        .assert<{ user: ErrorMessage }>(({ user }) => {
          expect(user.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the id is malformed', test => {
      test
        .arrange(() => ({ id: '1234' }))
        .act(fetchUser)
        .assert<{ user: ErrorMessage }>(({ user }) => {
          expect(user.code).toBe(400);
        });
    });

    suite.test('Will return the user object when there is an existing user', test => {
      test
        .arrange(createUser)
        .arrange<CreateUserOutput, CreateUserOutput & UserInput>(({ created }) => ({ id: created.id, created }))
        .act(fetchUser)
        .assert<UserOutput & CreateUserOutput>(({ created, user }) => {
          expect(user).toEqual(created);
        });
    });
  });

  suite.suite('The POST user endpoint', suite => {
    suite.afterEach(async () => {
      const users: User[] = await fetch(`http://localhost:${port}/users`).then(response => response.json());
      await cleanUsers({ users });
    });

    suite.test('Will return the user object once created', test => {
      test.act(createUser).assert<CreateUserOutput>(({ created }) => {
        expect(created).toBeDefined();
      });
    });

    suite.test('Will trigger a 400 if the values are invalid', test => {
      test
        .arrange(() => ({ username: null }))
        .act(createUser)
        .assert<{ created: ErrorMessage }>(({ created }) => {
          expect(created.code).toBe(400);
        });
    });
  });

  suite.suite('The PATCH user by id endpoint', suite => {
    suite.afterEach(async () => {
      const users: User[] = await fetch(`http://localhost:${port}/users`).then(response => response.json());
      await cleanUsers({ users });
    });

    suite.test('Will return a 404 when there are no users', test => {
      test
        .arrange(() => ({ user: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', username: 'test' } }))
        .act(updateUser)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the user does not exist', test => {
      test
        .arrange(createUser)
        .arrange(() => ({ user: { id: '8aec7349-5d4f-4dce-b576-182841348a3e', username: 'test' } }))
        .act(updateUser)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the id is malformed', test => {
      test
        .arrange(() => ({ user: { id: '1234', username: 'test' } }))
        .act(updateUser)
        .assert<{ updated: ErrorMessage }>(({ updated }) => {
          expect(updated.code).toBe(400);
        });
    });

    suite.test('Will return the updated user object when there is an existing user', test => {
      const newUsername = 'foo';

      test
        .arrange(createUser)
        .arrange<CreateUserOutput, CreateUserOutput & UserOutput>(({ created }) => ({
          user: { ...created, username: newUsername },
          created,
        }))
        .act(updateUser)
        .assert<CreateUserOutput & UpdateUserOutput>(({ created, updated }) => {
          expect(updated.id).toBe(created.id);
          expect(updated.username).toBe(newUsername);
          expect(updated.createdAt).toBe(created.createdAt);
        });
    });
  });

  suite.suite('The DELETE user by id endpoint', suite => {
    suite.afterEach(async () => {
      const users: User[] = await fetch(`http://localhost:${port}/users`).then(response => response.json());
      await cleanUsers({ users });
    });

    suite.test('Will return a 404 when there are no users', test => {
      test
        .arrange(() => ({ user: { id: '8aec7349-5d4f-4dce-b576-182841348a3e' } }))
        .act(deleteUser)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(404);
        });
    });

    suite.test('Will return a 404 when the user does not exist', test => {
      test
        .arrange(createUser)
        .arrange(() => ({ user: { id: '8aec7349-5d4f-4dce-b576-182841348a3e' } }))
        .act(deleteUser)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(404);
        });
    });

    suite.test('Will return a 400 when the id is malformed', test => {
      test
        .arrange(() => ({ user: { id: '1234' } }))
        .act(deleteUser)
        .assert<{ deleted: ErrorMessage }>(({ deleted }) => {
          expect(deleted.code).toBe(400);
        });
    });

    suite.test('Will return the deleted user object when there is an existing user', test => {
      test
        .arrange(createUser)
        .arrange<CreateUserOutput, CreateUserOutput & UserOutput>(({ created }) => ({
          user: created,
          created,
        }))
        .act(deleteUser)
        .act(fetchUsers)
        .assert<CreateUserOutput & DeleteUserOutput & UsersOutput>(({ created, deleted, users }) => {
          expect(deleted).toEqual(created);

          expect(users).not.toContain(deleted);
        });
    });
  });
});
