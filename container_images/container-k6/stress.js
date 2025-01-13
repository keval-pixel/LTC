import http from 'k6/http';
import { check, sleep } from 'k6';

// Test configuration
export let options = {
    vus: 1, // Number of virtual users
    duration: "1m", // Test duration
};

// Base URL of the Gerrit instance
const BASEURL = 'http://192.168.193.148:8080';
// Helper function to authenticate (if needed)
function stress_request() {
    let authUrl = `${BASEURL}/load?cpu=2&memory=2`;
    let res = http.get(authUrl);
    check(res, {
        'request successfull': (r) => r.status === 200,
    });
}


// Main test function
export default function () {
    // Authenticate if needed
    stress_request();

    // Sleep to simulate think time between user actions
    sleep(1);
}
