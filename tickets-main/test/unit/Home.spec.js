import Home from '@/views/Home'
import Carousel from '@/components/Carousel'
import Films from '@/components/Films'
import Info from '@/components/Info'
import Footer from '@/components/Footer'
import { shallowMount } from '@vue/test-utils'
import { expect } from 'chai'

describe('Home', () => {
  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(Home);
  });

  it('is called Home', () => {
    expect(wrapper.name()).to.equal('home');
  });

  it('should render Carousel', () => {
    expect(wrapper.find(Carousel).exists()).to.be.true;
  });

  it('should render Films on mount', () => {
    expect(wrapper.contains(Films)).to.be.true
  });

  it('should render Info', () => {
    expect(wrapper.find(Info).exists()).to.be.true;
  });

  it('should render Footer', () => {
    expect(wrapper.find(Footer).exists()).to.be.true;
  });
});