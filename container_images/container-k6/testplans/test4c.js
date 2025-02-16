import http from 'k6/http';
import { check, sleep } from 'k6';

// Test configuration
export const options = {
  scenarios: {
    constant_request_rate: {
      executor: 'constant-arrival-rate',
      rate: 1,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 5,
      maxVUs: 5,
    },
  },
};
// Base URL of the Gerrit instance
const BASEURL = 'http://192.168.193.148:30306';
// Helper function to authenticate (if needed)
function authenticate() {
    let authUrl = `${BASEURL}/a/accounts/self/detail`;
    let res = http.get(authUrl, {
        headers: {
            'Authorization': 'Basic YWRtaW46YWRtaW4=', // Correct Authorization header
            'Accept': 'application/json', // Optional: Set Accept header
        },
    });
    check(res, {
        'authenticated successfully': (r) => r.status === 200,
    });
}

// Function to create a new change
function createChange() {
    let url = `${BASEURL}/a/changes/`;
    let payload = JSON.stringify({
        project: "testing", // Replace with your project name
        branch: "master", // Replace with the target branch
        subject: "K6 Test Commit",
        topic: "test-topic",
    });
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.post(url, payload, params);
    check(res, {
        'change created successfully': (r) => r.status === 201,
    });

    //clean the initial closing characters that protect the json against code injection
    let cleanres = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();
    // Extract the change ID from the response
    let changeId = JSON.parse(cleanres).change_id;
    return changeId;
}

// Function to commit a change (add a file and commit)
function commitChange(changeId) {
    const filename = `File_${Date.now()}${__VU}`;
    let url = `${BASEURL}/a/changes/${changeId}/edit/${filename}`;
    let payload = JSON.stringify({
        binary_content: "data:text/plain;base64,SzYgZ2VuZXJhdGVkIGZpbGUK",
        file_mode: 100755// Convert content to Base64
    });
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.put(url, payload, params);
    check(res, {
        'file added successfully': (r) => r.status === 204,
    });

    // Publish the edit as a new patchset
    params = {
        headers: {
            'Authorization': 'Basic YWRtaW46YWRtaW4='
        },
    };

    let publishUrl = `${BASEURL}/a/changes/${changeId}/edit:publish`;
    let publishRes = http.post(publishUrl, null, params);
    check(publishRes, {
        'change published successfully': (r) => r.status === 200 || r.status === 204,
    });

    // Request to get the change details
    let changeDetailsUrl = `${BASEURL}/a/changes/${changeId}/?o=ALL_REVISIONS`
    let changeDetails = http.get(changeDetailsUrl, params);

    // Check if the request was successful
    check(changeDetails, {
        'Change details fetched successfully': (r) => r.status === 200,
    });

    let cleanres = changeDetails.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();
    // Extract the revision ID from the publish response
    let revisionId = JSON.parse(cleanres).current_revision;
    return revisionId;
}

// Function to review a change
function reviewChange(changeId, revisionId) {
    let url = `${BASEURL}/a/changes/${changeId}/revisions/${revisionId}/review`;
    let payload = JSON.stringify({
        message: "Looks good to me! Approved",
        labels: {
            "Code-Review": 2,
        },
    });
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.post(url, payload, params);
    check(res, {
        'reviewed change successfully': (r) => r.status === 200,
    });
}

// Function to submit a change
function submitChange(changeId) {
    let url = `${BASEURL}/a/changes/${changeId}/submit`;

    let params = {
        headers: {
            'Authorization': 'Basic YWRtaW46YWRtaW4=', // Correct Authorization header
            'Accept': 'application/json', // Optional: Set Accept header
        },
    };
    
    let res = http.post(url, null, params);
    check(res, {
        'submitted change successfully': (r) => r.status === 200,
    });
}

// Main test function
export default function () {
    // Authenticate if needed
    authenticate();

    // Step 1: Create a new change
    let changeId = createChange();
    console.log(`Change ID: ${changeId}`);

    // Step 2: Commit a change (create a patchset)
    let revisionId = commitChange(changeId);
    console.log(`Revision ID: ${revisionId}`);

    // Step 3: Review the change
    reviewChange(changeId, revisionId);

    // Step 4: Submit the change
    // submitChange(changeId);

    // Sleep to simulate think time between user actions
    // sleep(1);
}
