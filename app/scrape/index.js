const puppeteer = require('puppeteer');
const fs = require('fs');

//Start the browser and create a browser instance
let browserInstance = startBrowser();
scrapeAll(browserInstance);

async function startBrowser() {
    let browser;
    try {
        console.log("Opening the browser......");
        browser = await puppeteer.launch({
            headless: false,
            args: ["--disable-setuid-sandbox"],
            'ignoreHTTPSErrors': true
        });
    } catch (err) {
        console.log("Could not create a browser instance => : ", err);
    }
    return browser;
}

async function scrapeAll(browserInstance) {
    let browser;
    try {
        browser = await browserInstance;
        let scrapedData = {};
        // Call the scraper for different set of books to be scraped
        scrapedData = await scraperObject.scraper(browser);
        await browser.close();
        fs.writeFile("data.json", JSON.stringify(scrapedData), 'utf8', function (err) {
            if (err) {
                return console.log(err);
            }
            console.log("The data has been scraped and saved successfully! View it at './data.json'");
        });
    }
    catch (err) {
        console.log("Could not resolve the browser instance => ", err);
    }
}

const scraperObject = {
    url: 'https://nihongopicnic.notion.site/Nihongo-Picnic-Podcast-s-Transcript-e6c923a2d9f34c1fa278bb5e4531ea0f',
    async scraper(browser) {
        let page = await browser.newPage();
        console.log(`Navigating to ${this.url}...`);
        // Navigate to the selected page
        await page.goto(this.url, { waitUntil: 'networkidle0' });

        for (;;) {
            const isMore = await page.evaluate(() => {
                let eles = document.getElementsByClassName('arrowDown');
                if (eles.length == 0) {
                    return false;
                }
                eles[0].parentElement.click();
                return true;
            });
            if (!isMore) {
                break;
            }
        }        

        let episodes = await page.$$eval('.notion-table-view-row', rows => {
            
            //return row.querySelectorAll('span').map(el => el.textContent);
            rows = rows.map(el => {
                let text = Array.from(el.querySelectorAll('.notion-table-view-cell')).map(element => {
                    return element.textContent;
                });
                let links = Array.from(el.querySelectorAll('.notion-table-view-cell a')).map(element => {
                    return element.href;
                });
                return {
                    epNumber: text[0],
                    title: text[1],
                    date: text[2],
                    level: text[3].trim(),
                    link: links[0]
                };
            });
            return rows;
        });
        console.log('episode count', episodes.length);

        let scrapedData = [];

        let episodePromise = (ep) => new Promise(async (resolve, reject) => {
            let newPage = await browser.newPage();
            await newPage.goto(ep.link,{ waitUntil: 'networkidle0' });
            ep.content = await newPage.$eval('.notion-page-content', text => text.textContent);
            resolve(ep);
            await newPage.close();
        });

        episodes = episodes.slice(0, 3);
        console.log('eps', episodes.length);
        
        for (let i = 0; i < episodes.length; i++) {
            scrapedData[i] = await episodePromise(episodes[i]);
        }

        page.close();

        return scrapedData;

        // const [response] = await Promise.all([
        //     page.waitForNavigation(waitOptions),
        //     page.click(selector, clickOptions),
        //   ]);



        // Wait for the required DOM to be rendered
        async function scrapeCurrentPage() {

            // await page.waitForSelector('.page_inner');
            // // Get the link to all the required books
            // let urls = await page.$$eval('section ol > li', links => {
            //     // Make sure the book to be scraped is in stock
            //     links = links.filter(link => link.querySelector('.instock.availability > i').textContent !== "In stock")
            //     // Extract the links from the data
            //     links = links.map(el => el.querySelector('h3 > a').href)
            //     return links;
            // });

            // Loop through each of those links, open a new page instance and get the relevant data from them
            let pagePromise = (link) => new Promise(async (resolve, reject) => {
                let dataObj = {};
                let newPage = await browser.newPage();
                await newPage.goto(link);
                dataObj['bookTitle'] = await newPage.$eval('.product_main > h1', text => text.textContent);
                dataObj['bookPrice'] = await newPage.$eval('.price_color', text => text.textContent);
                dataObj['noAvailable'] = await newPage.$eval('.instock.availability', text => {
                    // Strip new line and tab spaces
                    text = text.textContent.replace(/(\r\n\t|\n|\r|\t)/gm, "");
                    // Get the number of stock available
                    let regexp = /^.*\((.*)\).*$/i;
                    let stockAvailable = regexp.exec(text)[1].split(' ')[0];
                    return stockAvailable;
                });
                dataObj['imageUrl'] = await newPage.$eval('#product_gallery img', img => img.src);
                dataObj['bookDescription'] = await newPage.$eval('#product_description', div => div.nextSibling.nextSibling.textContent);
                dataObj['upc'] = await newPage.$eval('.table.table-striped > tbody > tr > td', table => table.textContent);
                resolve(dataObj);
                await newPage.close();
            });

            for (link in urls) {
                let currentPageData = await pagePromise(urls[link]);
                scrapedData.push(currentPageData);
                // console.log(currentPageData);
            }
            // When all the data on this page is done, click the next button and start the scraping of the next page
            // You are going to check if this button exist first, so you know if there really is a next page.
            let nextButtonExist = false;
            try {
                const nextButton = await page.$eval('.next > a', a => a.textContent);
                nextButtonExist = true;
            }
            catch (err) {
                nextButtonExist = false;
            }
            if (nextButtonExist) {
                await page.click('.next > a');
                return scrapeCurrentPage(); // Call this function recursively
            }
            await page.close();
            return scrapedData;
        }
        // let data = await scrapeCurrentPage();
        // console.log(data);
        data = 'foo'
        return data;
    }
}

