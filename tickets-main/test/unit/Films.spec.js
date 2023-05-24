import Films from '@/components/Films'
import Modal from '@/components/Modal'
import { BDropdown } from 'bootstrap-vue'
import { shallowMount } from '@vue/test-utils'
import { expect } from 'chai'

describe('Films', () => {
  it('has correct props and attributes', () => {

    const cities = [
      {
        "id": "moskva",
        "name": "Москва"
      },
      {
        "id": "tumen",
        "name": "Тюмень"
      }
    ]

    const wrapper = shallowMount(Films,{
      data() {
        return {
          cities
        }
      },
    })

    expect(wrapper.find('#city-0').text()).to.equal('Москва')
    expect(wrapper.find('#city-1').text()).to.equal('Тюмень')
    // console.log(wrapper.html())
  })
})

describe('Films', () => {
  let wrapper;

  beforeEach(() => {
    wrapper = shallowMount(Films);
  });
 
  it('is called Films', () => {
    expect(wrapper.name()).to.equal('films');
  });

  it('should render BDropdown', () => {
    expect(wrapper.contains(BDropdown)).to.be.true
  })

  it('has a dropdown with id cities-dropdown', () => {
    expect(wrapper.contains('#cities-dropdown')).to.be.true
  })

  it('should render Modal', () => {
    expect(wrapper.contains(Modal)).to.be.true
  })

  it('has correct header', () => {
    expect(wrapper.find('h3').text()).to.equal('Фильмы')
  })
})