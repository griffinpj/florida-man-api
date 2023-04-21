import fs from 'fs/promises';
import fetch from 'node-fetch';

const format = (date) => new Date(date).toLocaleDateString('en-US', { year: 'numeric', month: '2-digit', day: '2-digit' }).replace(/\//g, '-')

// Define the base URL of the API endpoint
const baseUrl = 'http://localhost:8080/v1/search?date=';

// Define the start and end dates
const startDate = new Date('01-01-2000');
const endDate = new Date('12-31-2000');

// Define the delay between each request (in milliseconds)
const delay = 20000;

// Define a function to send the request for the current date
async function sendRequest(currentDate) {
  // Build the URL for the current date
  const url = `${baseUrl}${format(currentDate, 'mm-dd-yyyy')}`;

  try {
    // Send a GET request to the URL and wait for the response
    const response = await fetch(url);
    
    // Stop the loop and print the current date if the response is null
    if (response.status === 204) {
      console.log(`No data for ${format(currentDate, 'mm-dd-yyyy')}`);
      return;
    }

    // Parse the response as JSON and write it to a file
    const jsonData = await response.json();
      console.log(jsonData);
      console.log(url);
    const filename = `saved/${format(currentDate, 'mm-dd-yyyy')}.json`;
    await fs.writeFile(filename, JSON.stringify(jsonData));
    console.log(`Saved data to ${filename}`);
  } catch (err) {
    console.error(err);
    console.log(`Error on ${format(currentDate, 'mm-dd-yyyy')}`);
  }
}

// Define a function to loop over each date with a delay
async function loopWithDelay(startDate, endDate, delay) {
  let currentDate = new Date(startDate);
  let hasError = false;

  while (currentDate <= endDate && !hasError) {
    // Send the request for the current date and wait for it to finish
    await sendRequest(currentDate);
    
    // Move to the next date
    currentDate.setDate(currentDate.getDate() + 1);

    // Wait for the specified delay before sending the next request
    if (currentDate <= endDate) {
      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  }
}

// Call the loop function with the start and end dates and delay
loopWithDelay(startDate, endDate, delay);
