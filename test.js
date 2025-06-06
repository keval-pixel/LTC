import http from 'k6/http';
import { check, sleep } from 'k6';

// Test configuration
export let options = {
    vus: 1, // Number of virtual users
//    duration: "1m", // Test duration
    Iterations: 1
};

// Base URL of the Gerrit instance
const BASEURL = 'http://192.168.122.200:8080';
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

function list_projects() {
    let authUrl = `${BASEURL}/projects/`;
    let res = http.get(authUrl, {
        headers: {
            'Authorization': 'Basic YWRtaW46YWRtaW4=', // Correct Authorization header
            'Accept': 'application/json', // Optional: Set Accept header
        },
    });
    check(res, {
        'authenticated successfully': (r) => r.status === 200,
    });
    // Remove the unwanted prefix
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();;

    // Parse the cleaned response body as JSON
    let jsonData;
    try {
      jsonData = JSON.parse(cleanedBody);
    } catch (e) {
      console.error('Error parsing JSON:', e);
    }
    return jsonData;
}

export function getChanges(project = null, status = null) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    if (project) {
      params['project'] = project;
    }
    if (status) {
      params['status'] = status;
    }
    let res = http.get(`${BASEURL}/a/changes/`, params);
    // Remove the unwanted prefix                             
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();;    
                                                                                      
    // Parse the cleaned response body as JSON                          
    let jsonData;                                       
    try {                                                    
      jsonData = JSON.parse(cleanedBody);   
    } catch (e) {                                             
      console.error('Error parsing JSON:', e);
    }
    return jsonData;
}

export function getChange(changeId) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/changes/${changeId}`, params);
    // Remove the unwanted prefix                             
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();;    
                                                                                      
    // Parse the cleaned response body as JSON                          
    let jsonData;                                       
    try {                                                    
      jsonData = JSON.parse(cleanedBody);   
    } catch (e) {                                             
      console.error('Error parsing JSON:', e);
    }
    return jsonData;
}

export function getChangeComments(changeId) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/changes/${changeId}/comments`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();              
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
}

export function getChangeRevisions(changeId) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/changes/${changeId}/revisions`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();;              
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
}

export function getChangeRevisionDetail(changeId, revisionId) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/changes/${changeId}/revisions/${revisionId}`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();              
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
}

export function getChangeFiles(changeId) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/changes/${changeId}/files`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();          
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
}

export function getChangeDrafts(changeId) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/changes/${changeId}/drafts`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();        
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
}

export function getChangeLabels(changeId) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/changes/${changeId}/labels`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();           
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
}

export function getProjects() {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/projects/`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();            
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
}

export function getProjectDetails(projectName) {
    let params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Basic YWRtaW46YWRtaW4=',
        },
    };
    let res = http.get(`${BASEURL}/a/projects/${projectName}`, params);
    // Remove the unwanted prefix                                                     
    const cleanedBody = res.body.replace(")]}'", "").replace(/\r?\n|\r/g, "").trim();              
                                                           
    // Parse the cleaned response body as JSON               
    let jsonData;                                                       
    try {                                                               
      jsonData = JSON.parse(cleanedBody);                               
    } catch (e) {                                                       
      console.error('Error parsing JSON:', e);                          
    }                                                                   
    return jsonData;
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
    let res = http.post(payload, params);
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
    let res = http.put(payload, params);
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
    let res = http.post(payload, params);
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
    
    let res = http.post(null, params);
    check(res, {
        'submitted change successfully': (r) => r.status === 200,
    });
}

// Main test function
export default function () {
    // Authenticate if needed
    authenticate();

    // Step 1: Create a new change
    const changes = getChanges('testing1', 'new');
    console.log(changes);

    // Example: Fetch a specific change by ID
    const changeId = changes[0].id;

    const changeDetails = getChange(changeId);
    console.log(changeDetails);

    // Example: Fetch comments for a specific change
    const comments = getChangeComments(changeId);
    console.log(comments);

    const revisions = getChangeRevisions(changeId);
    console.log(revisions);

    // Sleep to simulate think time between user actions
    sleep(1);
}
