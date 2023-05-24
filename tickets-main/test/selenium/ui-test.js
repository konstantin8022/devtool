require('chromedriver');
const assert = require('assert');
const {Builder, Key, By, until} = require('selenium-webdriver');

describe('Checkout booking notifications', function () {
  let driver;

  before(async function() {
    driver = await new Builder().forBrowser('chrome').build();
  });

  it('Try to book without seats', async function() {
    await driver.get('http://localhost:8080/');

    await driver.sleep(500);

    await driver.wait(until.elementLocated(By.id('cities-dropdown')), 5000);
    await driver.findElement(By.id('cities-dropdown')).click();

    await driver.wait(until.elementLocated(By.id('city-1')), 5000);
    await driver.findElement(By.id('city-1')).click();

    await driver.wait(until.elementLocated(By.id('movie-0')), 10000);
    await driver.findElement(By.id('movie-0')).click();

    await driver.wait(until.elementLocated(By.id('seances-dropdown')), 5000);
    await driver.findElement(By.id('seances-dropdown')).click();

    await driver.wait(until.elementLocated(By.id('seance-2')), 5000);
    await driver.findElement(By.id('seance-2')).click();

    await driver.wait(until.elementLocated(By.id('cinema-hall')), 5000);

    await driver.findElement(By.id('email-field')).click();
    await driver.findElement(By.id('email-field')).sendKeys('test@mail.com');

    await driver.findElement(By.id('submit-button')).click();

    await driver.wait(until.elementLocated(By.className('b-toast-danger')), 5000);

    await driver.sleep(500);

    let text = await driver.findElement(By.className('toast-body')).getText();

    assert.equal(text, 'Информация о заказе заполнена некорректно');
  });

  it('clean form and try to book without email', async function() {
    await driver.findElement(By.id('clean-form')).click();

    await driver.wait(until.elementLocated(By.id('seances-dropdown')), 5000);
    await driver.findElement(By.id('seances-dropdown')).click();

    await driver.wait(until.elementLocated(By.id('seance-2')), 5000);
    await driver.findElement(By.id('seance-2')).click();

    await driver.wait(until.elementLocated(By.id('cinema-hall')), 5000);

    await driver.wait(until.elementLocated(By.className('seat')), 5000);
    await driver.findElement(By.className('seat')).click();

    await driver.findElement(By.id('submit-button')).click();

    await driver.wait(until.elementLocated(By.className('b-toast-danger')), 5000);

    await driver.sleep(500);

    let text = await driver.findElement(By.className('toast-body')).getText();

    assert.equal(text, 'Информация о заказе заполнена некорректно');
  });

  it('clean form and try to book with incorrect email', async function() {
    await driver.findElement(By.id('clean-form')).click();

    await driver.wait(until.elementLocated(By.id('seances-dropdown')), 5000);
    await driver.findElement(By.id('seances-dropdown')).click();

    await driver.wait(until.elementLocated(By.id('seance-2')), 5000);
    await driver.findElement(By.id('seance-2')).click();

    await driver.wait(until.elementLocated(By.id('cinema-hall')), 5000);

    await driver.wait(until.elementLocated(By.className('seat')), 5000);
    await driver.findElement(By.className('seat')).click();

    await driver.findElement(By.id('email-field')).click();
    await driver.findElement(By.id('email-field')).sendKeys('test@mail');

    await driver.findElement(By.id('submit-button')).click();

    await driver.wait(until.elementLocated(By.className('b-toast-danger')), 5000);

    await driver.sleep(500);

    let text = await driver.findElement(By.className('toast-body')).getText();

    assert.equal(text, 'Информация о заказе заполнена некорректно');
  });

  it('clean form and try to book without seats and email', async function() {
    await driver.findElement(By.id('clean-form')).click();

    await driver.wait(until.elementLocated(By.id('seances-dropdown')), 5000);
    await driver.findElement(By.id('seances-dropdown')).click();

    await driver.wait(until.elementLocated(By.id('seance-2')), 5000);
    await driver.findElement(By.id('seance-2')).click();

    await driver.wait(until.elementLocated(By.id('cinema-hall')), 5000);

    await driver.findElement(By.id('submit-button')).click();

    await driver.wait(until.elementLocated(By.className('b-toast-danger')), 5000);

    await driver.sleep(500);

    let text = await driver.findElement(By.className('toast-body')).getText();

    assert.equal(text, 'Информация о заказе заполнена некорректно');
  });

  it('close modal', async function() {
    await driver.findElement(By.id('close-modal')).click();

    await driver.wait(until.elementLocated(By.className('b-toast-primary')), 5000);

    await driver.sleep(500);

    let text = await driver.findElement(By.className('toast-body')).getText();

    assert.equal(text, 'Ваш заказ был отменён');
  });

  after(() => driver && driver.quit());
})