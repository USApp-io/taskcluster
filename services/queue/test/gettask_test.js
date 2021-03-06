const slugid = require('slugid');
const taskcluster = require('taskcluster-client');
const assume = require('assume');
const helper = require('./helper');
const testing = require('taskcluster-lib-testing');

helper.secrets.mockSuite(testing.suiteName(), ['aws', 'azure'], function(mock, skipping) {
  helper.withAmazonIPRanges(mock, skipping);
  helper.withPulse(mock, skipping);
  helper.withS3(mock, skipping);
  helper.withQueueService(mock, skipping);
  helper.withEntities(mock, skipping);
  helper.withServer(mock, skipping);

  const taskDef = {
    provisionerId: 'no-provisioner-extended-extended',
    workerType: 'test-worker-extended-extended',
    schedulerId: 'my-scheduler-extended-extended',
    taskGroupId: 'dSlITZ4yQgmvxxAi4A8fHQ',
    dependencies: [],
    requires: 'all-completed',
    routes: [],
    priority: 'lowest',
    retries: 5,
    created: taskcluster.fromNowJSON(),
    deadline: taskcluster.fromNowJSON('3 days'),
    expires: taskcluster.fromNowJSON('3 days'),
    scopes: [],
    payload: {},
    metadata: {
      name: 'Unit testing task',
      description: 'Task created during unit tests',
      owner: 'jonsafj@mozilla.com',
      source: 'https://github.com/taskcluster/taskcluster-queue',
    },
    tags: {
      purpose: 'taskcluster-testing',
    },
    extra: {},
  };

  test('task(taskId) is correct', async () => {
    const taskId = slugid.v4();

    await helper.queue.createTask(taskId, taskDef);
    const taskDef2 = await helper.queue.task(taskId);
    assume(taskDef2).deep.equals(taskDef);
  });

  test('task(taskId) doesn\'t require credentials', async () => {
    const taskId = slugid.v4();
    await helper.queue.createTask(taskId, taskDef);

    helper.scopes('none');
    const taskDef2 = await helper.queue.task(taskId);
    assume(taskDef2).deep.equals(taskDef);
  });
});
